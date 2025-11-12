# Releasing

## Automated Release Process

The release workflow is now fully automated. Simply update the VERSION file and push to main.

1. **Update VERSION file**

   ```bash
   echo "0.2.1" > VERSION
   ```

2. **Commit and push**

   ```bash
   git add VERSION
   git commit -m "chore: bump version to 0.2.1"
   git push origin main
   ```

3. **CI automatically:**
   - Detects VERSION file change
   - Reads version from VERSION file
   - Creates git tag `v0.2.1` (if it doesn't exist)
   - Pushes the tag
   - Runs tests and builds
   - Creates GitHub release

## Versioning Guidelines

- **Patch** (0.2.0 → 0.2.1): Bug fixes, small improvements
- **Minor** (0.2.0 → 0.3.0): New features, backward compatible
- **Major** (0.2.0 → 1.0.0): Breaking changes

The workflow only triggers when the VERSION file is changed in a push to main, preventing accidental releases.
