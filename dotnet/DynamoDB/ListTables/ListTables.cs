using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Threading;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace DynamoDBCRUD
{
    class ListTables
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }

        static async Task<ListTablesResponse> ShowTablesAsync(bool debug, IAmazonDynamoDB client)
        {
            var response = await client.ListTablesAsync(new ListTablesRequest { });

            return response;
        }

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("ListTables.exe [-r REGION] [-h]");
            Console.WriteLine("");
            Console.WriteLine(" REGION is optional, and defaults to us-west-2");
            Console.WriteLine(" -h prints this message and quits");
        }

        static void Main(string[] args)
        {
            bool debug = false;
            string region = "us-west-2";

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
                    case "-r":
                        i++;
                        region = args[i];
                        break;
                    default:
                        break;
                }

                i++;
            }

            DebugPrint(debug, "Debugging enabled\n");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);


            Task<ListTablesResponse> response = ShowTablesAsync(debug, client);

            Console.WriteLine("Found " + response.Result.TableNames.Count.ToString() + " tables in " + region + " region:");

            foreach (var table in response.Result.TableNames)
            {
                Console.WriteLine("  " + table);
            }
        }
    }
}
