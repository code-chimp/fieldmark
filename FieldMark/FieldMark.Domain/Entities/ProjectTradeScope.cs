namespace FieldMark.Domain.Entities;

public class ProjectTradeScope
{
    public Guid ProjectId { get; private set; }
    public Guid TradeTypeId { get; private set; }

    private ProjectTradeScope() { }
}
