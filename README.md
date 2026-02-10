# Judgeval Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/JudgmentLabs/judgeval-go.svg)](https://pkg.go.dev/github.com/JudgmentLabs/judgeval-go)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Installation

Find the latest version on [pkg.go.dev](https://pkg.go.dev/github.com/JudgmentLabs/judgeval-go).

**Go modules:**

```bash
go get github.com/JudgmentLabs/judgeval-go
```

## Usage

### Tracer

```go
package main

import (
    "context"
    "os"

    judgeval "github.com/JudgmentLabs/judgeval-go"
)

func main() {
    client, err := judgeval.NewJudgeval(
        "my-project",
        judgeval.WithAPIKey(os.Getenv("JUDGMENT_API_KEY")),
        judgeval.WithOrganizationID(os.Getenv("JUDGMENT_ORG_ID")),
    )
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    tracer, err := client.Tracer.Create(ctx, judgeval.TracerCreateParams{})
    if err != nil {
        panic(err)
    }
    defer tracer.Shutdown(ctx)

    _, span := tracer.Span(ctx, "my-operation")
    defer span.End()

    tracer.SetInput(span, "user input data")
    tracer.SetOutput(span, "operation result")
}
```

### Scorer

```go
package main

import (
    "context"
    "os"

    judgeval "github.com/JudgmentLabs/judgeval-go"
)

func main() {
    client, err := judgeval.NewJudgeval(
        "my-project",
        judgeval.WithAPIKey(os.Getenv("JUDGMENT_API_KEY")),
        judgeval.WithOrganizationID(os.Getenv("JUDGMENT_ORG_ID")),
    )
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    tracer, err := client.Tracer.Create(ctx, judgeval.TracerCreateParams{})
    if err != nil {
        panic(err)
    }
    defer tracer.Shutdown(ctx)

    scorer := client.Scorers.BuiltIn.AnswerCorrectness(judgeval.AnswerCorrectnessScorerParams{
        Threshold: judgeval.Float(0.7),
    })

    example := judgeval.NewExample(judgeval.ExampleParams{
        "input":           "What is 2+2?",
        "actual_output":   "4",
        "expected_output": "4",
    })

    spanCtx, span := tracer.Span(ctx, "evaluation")
    defer span.End()

    tracer.AsyncEvaluate(spanCtx, scorer, example)
}
```

## Documentation

- [API Documentation](https://pkg.go.dev/github.com/JudgmentLabs/judgeval-go)
- [Full Documentation](https://docs.judgmentlabs.ai/)

## License

Apache 2.0
