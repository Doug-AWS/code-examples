using System;
using System.Collections.Generic;
// using System.Linq;
// using System.Threading.Tasks;
using System.Net;
// using System.Web;

using Amazon.Lambda.APIGatewayEvents;
using Amazon.Lambda.Core;
using Newtonsoft.Json;

// Assembly attribute to enable the Lambda function's JSON input to be converted into a .NET class.
[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.SystemTextJson.DefaultLambdaJsonSerializer))]

namespace QuestionsFunction
{
    public class Function
    {
        void LogMessage(ILambdaContext ctx, string msg)
        {
            ctx.Logger.LogLine(
                string.Format("{0}:{1} - {2}",
                    ctx.AwsRequestId,
                    ctx.FunctionName,
                    msg));
        }

        APIGatewayProxyResponse CreateResponse(DateTime? result)
        {
            int statusCode = (result != null) ?
                (int)HttpStatusCode.OK :
                (int)HttpStatusCode.InternalServerError;

            string body = (result != null) ?
                JsonConvert.SerializeObject(result) : string.Empty;

            var response = new APIGatewayProxyResponse
            {
                StatusCode = statusCode,
                Body = body,
                Headers = new Dictionary<string, string>
                {
                    { "Content-Type", "application/json" },
                    { "Access-Control-Allow-Origin", "*" }
                }
            };

            return response;
        }

        /// <summary>
        /// A Lambda function that takes an API Gateway request
        /// </summary>
        /// <param name="request"></param>
        /// <param name="context"></param>
        /// <returns>APIGatewayProxyResponse</returns>
        public APIGatewayProxyResponse FunctionHandler(APIGatewayProxyRequest request, ILambdaContext context)
        {
            LogMessage(context, "Processing request started");

            APIGatewayProxyResponse response;
            try
            {
                var result = DateTime.UtcNow;
                response = CreateResponse(result);

                LogMessage(context, "Processing request succeeded.");
            }
            catch (Exception ex)
            {
                LogMessage(context, string.Format("Processing request failed - {0}", ex.Message));
                response = CreateResponse(null);
            }

            return response;
        }
    }
}
