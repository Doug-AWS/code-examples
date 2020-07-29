using Amazon.CDK;
using Amazon.CDK.AWS.APIGateway;
using Amazon.CDK.AWS.DynamoDB;
using Amazon.CDK.AWS.Lambda;

namespace SimpleWebService
{
    public class SimpleWebServiceStack : Stack
    {
        internal SimpleWebServiceStack(Construct scope, string id, IStackProps props = null) : base(scope, id, props)
        {
            var userTable = new Table(this, "questions", new TableProps {
                PartitionKey = new Attribute { 
                    Name = "id", 
                    Type = AttributeType.STRING 
                }
            });

            var api = new RestApi(this, "Simple Web service API", new RestApiProps {});

            var handler = new Function(this, "SimpleServiceApiHandler", new FunctionProps {
               Runtime = Runtime.DOTNET_CORE_3_1,
               Code = Code.FromAsset("resources"),
               Handler = "simpleWebService.handler"
            });

            // So Lambda function can access the table
            handler.AddEnvironment("TABLE_NAME", userTable.TableName);

            // So Lambda function can read/write to table
            userTable.GrantReadWriteData(handler);

            var apiHandler = new LambdaIntegration(handler);

            var user = api.Root.AddResource("user").AddResource("{id}");
            user.AddMethod("GET", apiHandler);
        }
    }
}
