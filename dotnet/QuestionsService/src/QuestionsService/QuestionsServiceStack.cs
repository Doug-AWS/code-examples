using Amazon.CDK;

namespace QuestionsService
{
    public class QuestionsServiceStack : Stack
    {
        internal QuestionsServiceStack(Construct scope, string id, IStackProps props = null) : base(scope, id, props)
        {
            new QuestionsService(this, "Questions");
        }
    }
}
