namespace FieldMark.Domain.Entities;

public class ProjectTradeScope
{
    public Guid ProjectId { get; private set; }
    public Guid TradeTypeId { get; private set; }

    private ProjectTradeScope() { }

    // Internal constructor for use by Project.Create factory only.
    internal ProjectTradeScope(Guid projectId, Guid tradeTypeId)
    {
        ProjectId = projectId;
        TradeTypeId = tradeTypeId;
    }
}
