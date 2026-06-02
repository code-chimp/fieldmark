namespace FieldMark.Domain.Exceptions;

public sealed class InvalidProjectTransitionException : Exception
{
    public InvalidProjectTransitionException(string message)
        : base(message)
    {
    }
}
