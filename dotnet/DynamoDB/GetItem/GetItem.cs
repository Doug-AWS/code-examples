using System;
using System.Collections.Generic;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace DynamoDBCRUD
{
    class GetItem
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("GetItem.exe ITEM-ID [-t TABLE] [-r REGION] [-h]");
            Console.WriteLine("");
            Console.WriteLine(" TABLE is optional, and defaults to CustomersOrdersProducts");
            Console.WriteLine(" REGION is optional, and defaults to us-west-2");
            Console.WriteLine(" -h prints this message and quits");
        }

        static async Task<QueryResponse> GetItemAsync(IAmazonDynamoDB client, string table, string id)
        {
            var response = await client.QueryAsync(new QueryRequest
            {
                TableName = table,
                KeyConditionExpression = "ID = :v_Id",
                ExpressionAttributeValues = new Dictionary<string, AttributeValue>
                {
                    { ":v_Id", new AttributeValue
                    { S = id }
                    }
                }
            });

            return response;
        }
                
        static void Main(string[] args)
        {
            bool debug = false;
            string region = "us-west-2";
            string table = "CustomersOrdersProducts";
            string id = "";

            int i = 0;
            while (i < args.Length)
            {
                switch (args[i])
                {
                    case "-h":
                        Usage();
                        return;
                    case "-d":
                        debug = true;
                        break;
                    case "-i":
                        i++;
                        id = args[i];
                        break;
                    case "-r":
                        i++;
                        region = args[i];
                        break;
                    case "-t":
                        i++;
                        table = args[i];
                        break;
                    default:
                        break;
                }

                i++;
            }

            if ((table == "") || (id == "") || (region == ""))
            {
                Console.WriteLine("You must supply a non-empty table name (-t TABLE), item (-i ITEM), and region (-r REGIoN)");
                return;
            }

            DebugPrint(debug, "Debugging enabled");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<QueryResponse> response =  GetItemAsync(client, table, id);
                        
            foreach (var item in response.Result.Items)
            {
                foreach (string attr in item.Keys)
                {
                    if (item[attr].S != null)
                    {
                        Console.WriteLine(attr + ": " + item[attr].S);
                    }
                    else if (item[attr].N != null)
                    {
                        Console.WriteLine(attr + ": " + item[attr].N.ToString());
                    }
                }

                Console.WriteLine("");
            }
        }
    }
}
