using Amazon.CDK;
using Amazon.CDK.AWS.APIGateway;
using Amazon.CDK.AWS.DynamoDB;
using Amazon.CDK.AWS.Lambda;
using Amazon.CDK.AWS.S3;
using System.Collections.Generic;

namespace QuestionsService
{
    public class QuestionsService : Construct
    {
        public QuestionsService(Construct scope, string id) : base(scope, id)
        {
            var table = new Table(this, "QuestionsTable", new TableProps
            {
                PartitionKey = new Attribute
                {
                    Name = "id",
                    Type = AttributeType.STRING
                },
                TableName = "QuestionsTable",
            });

            IBucket mybucket = Bucket.FromBucketName(this, "mybucket", "dougs-groovy-s3-bucket");
            
            var handler = new Function(this, "QuestionsHandler", new FunctionProps
            {
                Runtime = Runtime.DOTNET_CORE_3_1,
                Code = Code.FromBucket(mybucket, "QuestionsFunction.zip"),
                Handler = "QuestionsFunction::QuestionsFunction.Function::FunctionHandler",                
                Environment = new Dictionary<string, string>
                {
                    ["TABLE"] = table.TableName
                }
            });

            table.GrantReadWriteData(handler);

            var api = new RestApi(this, "Questions-API", new RestApiProps
            {
                RestApiName = "Questions Service",
                Description = "This service services questions."
            });

            var getQuestionsIntegration = new LambdaIntegration(handler, new LambdaIntegrationOptions
            {
                RequestTemplates = new Dictionary<string, string>
                {
                    ["application/json"] = "{ \"statusCode\": \"200\" }"
                }
            });

            api.Root.AddMethod("GET", getQuestionsIntegration);
        }
    }
}
