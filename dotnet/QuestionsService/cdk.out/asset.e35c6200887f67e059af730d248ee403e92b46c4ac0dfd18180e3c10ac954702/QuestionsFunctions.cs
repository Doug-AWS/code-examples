using Amazon.CDK;
using Amazon.CDK.AWS.Lambda;
using Amazon.Lambda.Core;
using System;
using System.Net;
using System.Web;

namespace QuestionsService
{
    class Questions
    {
        //readonly string BUCKET_NAME = System.Environment.GetEnvironmentVariable("BUCKET");
        //readonly string TABLE_NAME = System.Environment.GetEnvironmentVariable("TABLE");

        public string FunctionHandler(object source, ILambdaContext context)
        {
            // var request = System.Web.HttpUtility.ParseQueryString(e.ToString());
            // Just barf out the type of calling object
            return "WTF"; // source.GetType().ToString();
        }
    }
}