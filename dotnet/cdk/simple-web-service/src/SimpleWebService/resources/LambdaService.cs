using System.Threading.Tasks;
using Amazon.Lambda;
using Amazon.Lambda.Core;


using Amazon.Lambda.DynamoDBEvents;




namespace SimpleWebService
{
    public class LambdaWebService
    {
        // Called whenever a new record is added to a DynamoDB table
		public async Task FunctionHandler(DynamoDBEvent dynamoEvent, ILambdaContext context)
		{
		context.Logger.LogLine($"Beginning to process {dynamoEvent.Records.Count} records...");

		foreach (var record in dynamoEvent.Records)
		{
			context.Logger.LogLine($"Event ID: {record.EventID}");
			context.Logger.LogLine($"Event Name: {record.EventName}");

			var streamRecordJson = _dynamoDbWriter.SerializeStreamRecord(record.Dynamodb);
			context.Logger.LogLine($"DynamoDB Record:{streamRecordJson}");
			context.Logger.LogLine(streamRecordJson);

			var logEntry = new LogEntry
			{
				Message = $"Movie '{record.Dynamodb.NewImage["Title"].S}' processed by lambda",
				DateTime = DateTime.Now
			};
			await _sqsWriter.WriteLogEntryAsync(logEntry);
			await _dynamoDbWriter.PutLogEntryAsync(logEntry);
		}

		context.Logger.LogLine("Stream processing complete.");
	}

    }
}