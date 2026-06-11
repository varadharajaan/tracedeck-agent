# Policy Config

Policies are YAML files decoded into strongly typed Go structs. Unknown fields
fail validation. Collection modes, archive providers, email providers,
severities, and sensitive capabilities are enums backed by centralized
constants.

Generate the policy schema with:

```powershell
go run ./agent/cmd/tracedeck-agent schema --out ./docs/schema/policy-v1alpha1.schema.json
```
