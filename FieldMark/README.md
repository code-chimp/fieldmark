# FieldMark

## Notes:

Create migration example:

```shell
dotnet ef migrations add ThatThingWeDid -p FieldMark.Data -s FieldMark.Web
```

Run migration example:

```shell
dotnet ef database update -p FieldMark.Data -s FieldMark.Web
```
