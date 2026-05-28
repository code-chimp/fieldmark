namespace FieldMark.Domain.Entities.Reference;

public sealed class ViolationCategory
{
    public Guid Id { get; private set; }
    public string Code { get; private set; } = string.Empty;
    public string Name { get; private set; } = string.Empty;
    public Guid? TradeTypeId { get; private set; }
    public string DefaultSeverity { get; private set; } = string.Empty;
    public string? Description { get; private set; }
    public bool Active { get; private set; }

    private ViolationCategory() { }

    public ViolationCategory(
        string code,
        string name,
        Guid? tradeTypeId,
        string defaultSeverity,
        string? description,
        bool active
    )
    {
        Id = Guid.NewGuid();
        Code = code;
        Name = name;
        TradeTypeId = tradeTypeId;
        DefaultSeverity = defaultSeverity;
        Description = description;
        Active = active;
    }
}
