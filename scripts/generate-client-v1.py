#!/usr/bin/env python3

import json
import os
import re
import shutil
import sys
from typing import Any, Dict, List, Optional, Set
import httpx

HTTP_METHODS = {"GET", "POST", "PUT", "PATCH", "DELETE"}
SUCCESS_STATUS_CODES = {"200", "201"}
SCHEMA_REF_PREFIX = "#/components/schemas/"


def resolve_ref(ref: str) -> str:
    assert ref.startswith(
        SCHEMA_REF_PREFIX
    ), f"Reference must start with {SCHEMA_REF_PREFIX}"
    return ref.replace(SCHEMA_REF_PREFIX, "")


def to_camel_case(name: str) -> str:
    parts = name.replace("-", "_").split("_")
    return parts[0] + "".join(word.capitalize() for word in parts[1:])


def to_snake_case(name: str) -> str:
    name = re.sub("(.)([A-Z][a-z]+)", r"\1_\2", name)
    return re.sub("([a-z0-9])([A-Z])", r"\1_\2", name).lower()


def to_pascal_case(name: str) -> str:
    clean = name.replace("-", "_")
    if "_" in clean:
        parts = clean.split("_")
        return "".join(part[:1].upper() + part[1:] for part in parts if part)
    return clean[:1].upper() + clean[1:] if clean else clean


def to_struct_name(name: str) -> str:
    return to_pascal_case(name)


def get_method_name_from_path(path: str, method: str) -> str:
    clean_path = path.strip("/").replace("/", "_").replace("-", "_")
    camel_case = to_camel_case(clean_path)
    return camel_case[0].upper() + camel_case[1:]


def get_method_name_from_operation(
    operation: Dict[str, Any], path: str, method: str
) -> str:
    operation_id = operation.get("operationId")
    if operation_id:
        name = to_snake_case(operation_id)
        name = re.sub(r"^(get|post|put|patch|delete)_v1_", r"\1_", name)
        name = re.sub(r"_by_project_id", "", name)
        name = name.replace("-", "_")
        return to_pascal_case(name)

    name = re.sub(r"\{[^}]+\}", "", path)
    name = name.strip("/").replace("/", "_").replace("-", "_")
    name = re.sub(r"_+", "_", name).strip("_")
    if not name:
        return "Index"
    if name.startswith("v1_"):
        name = name[3:]
    elif name == "v1":
        name = "index"
    return to_pascal_case(name)


def get_query_parameters(operation: Dict[str, Any]) -> List[Dict[str, Any]]:
    return [
        {
            "name": param["name"],
            "required": param.get("required", False),
            "type": param.get("schema", {}).get("type", "string"),
        }
        for param in operation.get("parameters", [])
        if param.get("in") == "query"
    ]


def get_schema_from_content(content: Dict[str, Any]) -> Optional[str]:
    if "application/json" in content:
        schema = content["application/json"].get("schema", {})
        if not isinstance(schema, dict):
            return None
        if "$id" in schema:
            return to_pascal_case(schema["$id"])
        if "$ref" in schema:
            return resolve_ref(schema["$ref"])
    if "text/plain" in content:
        schema = content["text/plain"].get("schema", {})
        if not isinstance(schema, dict):
            return None
        if "$id" in schema:
            return to_pascal_case(schema["$id"])
        if "$ref" in schema:
            return resolve_ref(schema["$ref"])
    return None


def get_request_schema(operation: Dict[str, Any]) -> Optional[str]:
    request_body = operation.get("requestBody", {})
    return (
        get_schema_from_content(request_body.get("content", {}))
        if request_body
        else None
    )


def get_response_schema(operation: Dict[str, Any]) -> Optional[str]:
    responses = operation.get("responses", {})
    for status_code in SUCCESS_STATUS_CODES:
        if status_code in responses:
            result = get_schema_from_content(responses[status_code].get("content", {}))
            if result:
                return result
    return None


def extract_dependencies(
    schema: Dict[str, Any],
    schemas_by_id: Dict[str, Dict[str, Any]],
    visited: Optional[Set[str]] = None,
) -> Set[str]:
    if visited is None:
        visited = set()

    if not isinstance(schema, dict):
        return set()

    dependencies: Set[str] = set()

    if "$ref" in schema:
        return dependencies

    schema_id = schema.get("$id")
    if schema_id and schema_id in visited:
        return dependencies

    if schema_id:
        visited.add(schema_id)
        full_schema = schemas_by_id.get(schema_id, schema)
    else:
        full_schema = schema

    if "properties" in full_schema and isinstance(full_schema["properties"], dict):
        for prop_schema in full_schema["properties"].values():
            if isinstance(prop_schema, dict):
                if "$id" in prop_schema:
                    dep_id = prop_schema["$id"]
                    dependencies.add(dep_id)
                    dependencies.update(
                        extract_dependencies(prop_schema, schemas_by_id, visited)
                    )
                else:
                    dependencies.update(
                        extract_dependencies(prop_schema, schemas_by_id, visited)
                    )

    if "items" in full_schema:
        items_schema = full_schema["items"]
        if isinstance(items_schema, dict):
            if "$id" in items_schema:
                dep_id = items_schema["$id"]
                dependencies.add(dep_id)
                dependencies.update(
                    extract_dependencies(items_schema, schemas_by_id, visited)
                )
            else:
                dependencies.update(
                    extract_dependencies(items_schema, schemas_by_id, visited)
                )

    for union_key in ("anyOf", "oneOf", "allOf"):
        if union_key in full_schema and isinstance(full_schema[union_key], list):
            for item in full_schema[union_key]:
                if isinstance(item, dict):
                    if "$id" in item:
                        dep_id = item["$id"]
                        dependencies.add(dep_id)
                        dependencies.update(
                            extract_dependencies(item, schemas_by_id, visited)
                        )
                    else:
                        dependencies.update(
                            extract_dependencies(item, schemas_by_id, visited)
                        )

    if "additionalProperties" in full_schema and isinstance(
        full_schema["additionalProperties"], dict
    ):
        dependencies.update(
            extract_dependencies(
                full_schema["additionalProperties"], schemas_by_id, visited
            )
        )

    return dependencies


def collect_schemas_with_id(spec: Dict[str, Any]) -> Dict[str, Dict[str, Any]]:
    schemas_by_id: Dict[str, Dict[str, Any]] = {}

    def collect_from_value(value: Any) -> None:
        if isinstance(value, dict):
            if "$id" in value:
                schema_id = value["$id"]
                if schema_id not in schemas_by_id:
                    schema_without_id = {k: v for k, v in value.items() if k != "$id"}
                    schemas_by_id[schema_id] = schema_without_id
            if "$ref" not in value:
                for v in value.values():
                    collect_from_value(v)
        elif isinstance(value, list):
            for item in value:
                collect_from_value(item)

    collect_from_value(spec)
    return schemas_by_id


def find_used_schemas(
    spec: Dict[str, Any], schemas_by_id: Dict[str, Dict[str, Any]]
) -> Set[str]:
    used_schemas = set()

    for path, path_item in spec.get("paths", {}).items():
        for method, operation in path_item.items():
            if not isinstance(operation, dict):
                continue
            if method.upper() not in HTTP_METHODS:
                continue

            request_body = operation.get("requestBody", {})
            if request_body:
                for content in request_body.get("content", {}).values():
                    if "schema" in content:
                        schema = content["schema"]
                        if "$id" in schema:
                            schema_id = schema["$id"]
                            used_schemas.add(schema_id)
                            if schema_id in schemas_by_id:
                                used_schemas.update(
                                    extract_dependencies(
                                        schemas_by_id[schema_id], schemas_by_id
                                    )
                                )

            for response in operation.get("responses", {}).values():
                if not isinstance(response, dict):
                    continue
                for content in response.get("content", {}).values():
                    if "schema" in content:
                        schema = content["schema"]
                        if "$id" in schema:
                            schema_id = schema["$id"]
                            used_schemas.add(schema_id)
                            if schema_id in schemas_by_id:
                                used_schemas.update(
                                    extract_dependencies(
                                        schemas_by_id[schema_id], schemas_by_id
                                    )
                                )
                        else:
                            used_schemas.update(
                                extract_dependencies(schema, schemas_by_id)
                            )

    return used_schemas


def get_go_type(schema: Dict[str, Any]) -> str:
    if not isinstance(schema, dict):
        return "any"
    if "$ref" in schema:
        return to_struct_name(resolve_ref(schema["$ref"]))
    if "$id" in schema:
        return to_struct_name(schema["$id"])

    for union_key in ["anyOf", "oneOf", "allOf"]:
        if union_key in schema:
            union_schemas = schema[union_key]
            types = set()

            for union_schema in union_schemas:
                if union_schema.get("type") == "null":
                    types.add("null")
                else:
                    types.add(get_go_type(union_schema))

            non_null_types = types - {"null"}
            if len(non_null_types) == 1:
                return list(non_null_types)[0]
            else:
                return "any"

    schema_type = schema.get("type", "object")
    type_mapping = {
        "string": "string",
        "integer": "int",
        "number": "float64",
        "boolean": "bool",
        "object": "any",
    }

    if schema_type == "array":
        items = schema.get("items", {})
        return f"[]{get_go_type(items)}" if items else "[]any"

    return type_mapping.get(schema_type, "any")


def generate_struct(className: str, schema: Dict[str, Any]) -> str:
    lines = [
        "package models",
        "",
        "import (",
        '    "encoding/json"',
        ")",
        "",
        f"type {className} struct {{",
    ]

    if "properties" in schema:
        for field_name, property_schema in schema["properties"].items():
            go_type = get_go_type(property_schema)
            json_tag = f'json:"{field_name},omitempty"'

            lines.append(f"    {to_struct_name(field_name)} {go_type} `{json_tag}`")

        lines.append("")

    lines.extend(
        [
            '    AdditionalProperties map[string]any `json:"-"`',
            "}",
            "",
            f"func (m *{className}) UnmarshalJSON(data []byte) error {{",
            f"    type Alias {className}",
            "    aux := &struct {",
            "        *Alias",
            "    }{",
            "        Alias: (*Alias)(m),",
            "    }",
            "    if err := json.Unmarshal(data, &aux); err != nil {{",
            "        return err",
            "    }}",
            "    m.AdditionalProperties = make(map[string]any)",
            "    if err := json.Unmarshal(data, &m.AdditionalProperties); err != nil {{",
            "        return err",
            "    }}",
            "    return nil",
            "}",
            "",
            f"func (m {className}) MarshalJSON() ([]byte, error) {{",
            f"    type Alias {className}",
            "    aux := &struct {",
            "        *Alias",
            "    }{",
            "        Alias: (*Alias)(&m),",
            "    }",
            "",
            "    result := make(map[string]any)",
            "",
            "    mainBytes, err := json.Marshal(aux)",
            "    if err != nil {{",
            "        return nil, err",
            "    }}",
            "",
            "    if err := json.Unmarshal(mainBytes, &result); err != nil {{",
            "        return nil, err",
            "    }}",
            "",
            "    for k, v := range m.AdditionalProperties {{",
            "        result[k] = v",
            "    }}",
            "",
            "    return json.Marshal(result)",
            "}",
        ]
    )

    return "\n".join(lines)


def generate_type_definition(class_name: str, schema: Dict[str, Any]) -> str:
    schema_type = schema.get("type", "object")
    if schema_type == "array":
        items = schema.get("items", {})
        item_type = get_go_type(items) if isinstance(items, dict) else "any"
        return "\n".join(
            [
                "package models",
                "",
                f"type {class_name} []{item_type}",
            ]
        )
    if schema_type == "object" or "properties" in schema:
        return generate_struct(class_name, schema)
    return "\n".join(
        [
            "package models",
            "",
            f"type {class_name} {get_go_type(schema)}",
        ]
    )


def extract_path_params(path: str) -> List[Dict[str, Any]]:
    params = []
    for match in re.finditer(r"\{(\w+)\}", path):
        params.append(
            {"name": match.group(1), "required": True, "type": "string", "in": "path"}
        )
    return params


def generate_method_signature(
    method_name: str,
    request_type: Optional[str],
    path_params: List[Dict[str, Any]],
    query_params: List[Dict[str, Any]],
    response_type: str,
) -> str:
    params = []

    for param in path_params:
        param_name = to_camel_case(param["name"])
        params.append(f"{param_name} string")

    for param in query_params:
        if param["required"]:
            params.append(f"{param['name']} string")

    if request_type:
        params.append(f"payload *models.{request_type}")

    for param in query_params:
        if not param["required"]:
            params.append(f"{param['name']} *string")

    response_type_ref = (
        f"models.{response_type}" if response_type != "any" else response_type
    )
    return f"func (c *Client) {method_name}({', '.join(params)}) (*{response_type_ref}, error) {{"


def generate_method_body(
    method_name: str,
    path: str,
    method: str,
    request_type: Optional[str],
    path_params: List[Dict[str, Any]],
    query_params: List[Dict[str, Any]],
    response_type: str,
) -> str:
    response_type_ref = (
        f"models.{response_type}" if response_type != "any" else response_type
    )
    lines = []

    if path_params:
        path_fmt = path
        path_args = []
        for param in path_params:
            param_name = to_camel_case(param["name"])
            path_fmt = path_fmt.replace(f"{{{param['name']}}}", "%s")
            path_args.append(param_name)
        lines.append(f'    path := fmt.Sprintf("{path_fmt}", {", ".join(path_args)})')
    else:
        lines.append(f'    path := "{path}"')

    if query_params:
        lines.append("    queryParams := make(map[string]string)")
        for param in query_params:
            param_name = param["name"]
            if param["required"]:
                lines.append(f'    queryParams["{param_name}"] = {param_name}')
            else:
                lines.extend(
                    [
                        f"    if {param_name} != nil {{",
                        f'        queryParams["{param_name}"] = *{param_name}',
                        "    }",
                    ]
                )

    if query_params:
        lines.append("    url := c.buildURL(path, queryParams)")
    else:
        lines.append("    url := c.buildURL(path, nil)")

    if method in ["GET", "DELETE"]:
        lines.extend(
            [
                f'    req, err := http.NewRequest("{method}", url, nil)',
                "    if err != nil {",
                "        return nil, err",
                "    }",
                "    c.setHeaders(req)",
            ]
        )
    else:
        payload_expr = "payload" if request_type else "struct{}{}"
        lines.extend(
            [
                f"    jsonPayload, err := json.Marshal({payload_expr})",
                "    if err != nil {",
                "        return nil, err",
                "    }",
                f'    req, err := http.NewRequest("{method}", url, bytes.NewBuffer(jsonPayload))',
                "    if err != nil {",
                "        return nil, err",
                "    }",
                "    c.setHeaders(req)",
            ]
        )

    lines.extend(
        [
            "    resp, err := c.doRequest(req)",
            "    if err != nil {",
            "        return nil, err",
            "    }",
            "    defer resp.Body.Close()",
            "",
            "    if resp.StatusCode >= 400 {",
            "        body, _ := io.ReadAll(resp.Body)",
            '        return nil, fmt.Errorf("HTTP Error: %d - %s", resp.StatusCode, string(body))',
            "    }",
            "",
            f"    var result {response_type_ref}",
            "    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {",
            "        return nil, err",
            "    }",
            "    return &result, nil",
        ]
    )

    return "\n".join(lines)


def generate_client_class(methods: List[Dict[str, Any]]) -> str:
    lines = [
        "package api",
        "",
        "import (",
        '    "bytes"',
        '    "encoding/json"',
        '    "fmt"',
        '    "io"',
        '    "net/http"',
        '    "net/url"',
        "",
        '    "github.com/JudgmentLabs/judgeval-go/internal/api/models"',
        '    "github.com/JudgmentLabs/judgeval-go/logger"',
        ")",
        "",
        "type Client struct {",
        "    baseURL        string",
        "    apiKey         string",
        "    organizationID string",
        "    httpClient     *http.Client",
        "}",
        "",
        "func NewClient(baseURL, apiKey, organizationID string) *Client {",
        "    return &Client{",
        "        baseURL:        baseURL,",
        "        apiKey:         apiKey,",
        "        organizationID: organizationID,",
        "        httpClient:     &http.Client{},",
        "    }",
        "}",
        "",
        "func (c *Client) doRequest(req *http.Request) (*http.Response, error) {",
        '    logger.Debug("HTTP %s %s", req.Method, req.URL.String())',
        "    resp, err := c.httpClient.Do(req)",
        "    if err != nil {",
        '        logger.Debug("HTTP error: %v", err)',
        "        return nil, err",
        "    }",
        '    logger.Debug("HTTP %s %s -> %d", req.Method, req.URL.String(), resp.StatusCode)',
        "    return resp, nil",
        "}",
        "",
        "func (c *Client) buildURL(path string, queryParams map[string]string) string {",
        "    u, _ := url.Parse(c.baseURL + path)",
        "    if len(queryParams) > 0 {",
        "        q := u.Query()",
        "        for k, v := range queryParams {",
        "            q.Set(k, v)",
        "        }",
        "        u.RawQuery = q.Encode()",
        "    }",
        "    return u.String()",
        "}",
        "",
        "func (c *Client) setHeaders(req *http.Request) {",
        '    req.Header.Set("Content-Type", "application/json")',
        '    req.Header.Set("Authorization", "Bearer "+c.apiKey)',
        '    req.Header.Set("X-Organization-Id", c.organizationID)',
        "}",
        "",
        "func (c *Client) GetBaseURL() string {",
        "    return c.baseURL",
        "}",
        "",
        "func (c *Client) GetAPIKey() string {",
        "    return c.apiKey",
        "}",
        "",
        "func (c *Client) GetOrganizationID() string {",
        "    return c.organizationID",
        "}",
        "",
    ]

    for method_info in methods:
        signature = generate_method_signature(
            method_info["name"],
            method_info["request_type"],
            method_info["path_params"],
            method_info["query_params"],
            method_info["response_type"],
        )
        lines.append(signature)

        body = generate_method_body(
            method_info["name"],
            method_info["path"],
            method_info["method"],
            method_info["request_type"],
            method_info["path_params"],
            method_info["query_params"],
            method_info["response_type"],
        )
        lines.append(body)
        lines.append("}")
        lines.append("")

    return "\n".join(lines)


def generate_api_files(spec: Dict[str, Any]) -> None:
    schemas_by_id = collect_schemas_with_id(spec)

    models_dir = "internal/api/models"
    if os.path.exists(models_dir):
        print(f"Clearing existing models directory: {models_dir}", file=sys.stderr)
        shutil.rmtree(models_dir)

    os.makedirs(models_dir, exist_ok=True)

    print("Generating model structs...", file=sys.stderr)
    for schema_name in sorted(schemas_by_id.keys()):
        struct_name = to_struct_name(schema_name)
        model_struct = generate_type_definition(struct_name, schemas_by_id[schema_name])

        with open(f"{models_dir}/{struct_name.lower()}.go", "w") as f:
            f.write(model_struct)

        print(f"Generated model: {struct_name}", file=sys.stderr)

    include_prefixes = ["/v1", "/otel"]
    filtered_paths = {
        path: spec_data
        for path, spec_data in spec["paths"].items()
        if any(path.startswith(prefix) for prefix in include_prefixes)
    }

    methods = []
    for path, path_data in filtered_paths.items():
        for method, operation in path_data.items():
            if method.upper() in HTTP_METHODS:
                method_name = get_method_name_from_operation(
                    operation, path, method.upper()
                )
                request_schema = get_request_schema(operation)
                response_schema = get_response_schema(operation)
                path_params = extract_path_params(path)
                query_params = get_query_parameters(operation)

                print(
                    f"{method_name} {request_schema} {response_schema} {path_params} {query_params}",
                    file=sys.stderr,
                )

                method_info = {
                    "name": method_name,
                    "path": path,
                    "method": method.upper(),
                    "request_type": (
                        to_struct_name(request_schema) if request_schema else None
                    ),
                    "query_params": query_params,
                    "path_params": path_params,
                    "response_type": (
                        to_struct_name(response_schema)
                        if response_schema
                        else "any"
                    ),
                }
                methods.append(method_info)

    api_dir = "internal/api"
    os.makedirs(api_dir, exist_ok=True)

    client_class = generate_client_class(methods)
    with open(f"{api_dir}/client.go", "w") as f:
        f.write(client_class)
    print(f"Generated: {api_dir}/client.go", file=sys.stderr)


def main():
    spec_file = (
        sys.argv[1] if len(sys.argv) > 1 else "http://localhost:10001/openapi/json"
    )

    try:
        if spec_file.startswith("http"):
            with httpx.Client() as client:
                response = client.get(spec_file)
                response.raise_for_status()
                spec = response.json()
        else:
            with open(spec_file, "r") as f:
                spec = json.load(f)

        generate_api_files(spec)

    except Exception as e:
        print(f"Error generating API client: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
