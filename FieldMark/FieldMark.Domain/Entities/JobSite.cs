namespace FieldMark.Domain.Entities;

public class JobSite
{
    public Guid Id { get; private set; }
    public Guid ProjectId { get; private set; }
    public string Label { get; private set; } = string.Empty;
    public string? Address { get; private set; }

    private JobSite() { }
}
