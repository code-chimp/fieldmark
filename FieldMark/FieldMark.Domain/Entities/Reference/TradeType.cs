namespace FieldMark.Domain.Entities.Reference;

public sealed class TradeType
{
    public Guid Id { get; private set; }
    public string Code { get; private set; } = string.Empty;
    public string Name { get; private set; } = string.Empty;
    public string? Description { get; private set; }
    public bool Active { get; private set; }

    private TradeType() { }

    public TradeType(string code, string name, string? description, bool active)
    {
        Id = Guid.NewGuid();
        Code = code;
        Name = name;
        Description = description;
        Active = active;
    }
}
