using System.Text.Json;

namespace FieldMark.Domain.Entities.Reference;

public sealed class ComplianceRule
{
    public Guid Id { get; private set; }
    public string Code { get; private set; } = string.Empty;
    public string Name { get; private set; } = string.Empty;
    public string Description { get; private set; } = string.Empty;
    public string RuleKind { get; private set; } = string.Empty;
    public JsonDocument Parameters { get; private set; } = JsonDocument.Parse("{}");
    public bool Active { get; private set; }

    private ComplianceRule() { }

    public ComplianceRule(
        string code,
        string name,
        string description,
        string ruleKind,
        JsonDocument parameters,
        bool active
    )
    {
        Id = Guid.NewGuid();
        Code = code;
        Name = name;
        Description = description;
        RuleKind = ruleKind;
        Parameters = parameters;
        Active = active;
    }
}
