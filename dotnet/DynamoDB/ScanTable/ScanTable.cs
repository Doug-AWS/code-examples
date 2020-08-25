using System;
using System.Collections.Generic;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace DynamoDBCRUD
{
    class ScanTable
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
            Console.WriteLine("ScanTable.exe ITEM-ID [-t TABLE] [-r REGION] [-h]");
            Console.WriteLine("");
            Console.WriteLine(" TABLE is optional, and defaults to CustomersOrdersProducts");
            Console.WriteLine(" REGION is optional, and defaults to us-west-2");
            Console.WriteLine(" -h prints this message and quits");
        }

        static async Task<ScanResponse> ScanItemsAsync(IAmazonDynamoDB client, string table, string volume)
        {
            var response = await client.ScanAsync(new ScanRequest
            {
                TableName = table,
                ExpressionAttributeValues = new Dictionary<string, AttributeValue> {
                    {":val", new AttributeValue { N = volume }}
                },
                FilterExpression = "Product_Quantity < :val",
                ProjectionExpression = "ID, Product_Description, Product_Quantity"
            });

            return response;
        }

        static void Main(string[] args)
        {
            bool debug = false;
            string region = "us-west-2";
            string table = "CustomersOrdersProducts";
            string minimum = "";

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
                    case "-m":
                        i++;
                        minimum = args[i];
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

            if ((table == "") || (minimum == "") || (region == ""))
            {
                Console.WriteLine("You must supply a non-empty table name (-t TABLE) and region (-r REGIoN), and minimum quantity (-m NUMBER)");
                return;
            }

            int number;

            try
            {
                number = int.Parse(minimum);
            }
            catch (FormatException)
            {
                Console.WriteLine(minimum + " is not an integer");
                return;
            }

            DebugPrint(debug, "Debugging enabled");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<ScanResponse> response = ScanItemsAsync(client, table, minimum);

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
