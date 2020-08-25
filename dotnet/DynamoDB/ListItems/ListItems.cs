using System;
using System.Collections.Generic;
using System.IO;
using System.Reflection;
using System.Threading;
using System.Threading.Tasks;

using System.Text.Json;
using System.Text.Json.Serialization;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.DocumentModel;
using Amazon.DynamoDBv2.Model;
using System.Diagnostics;
using System.Text;

namespace DynamoDBCRUD
{
    class ListItems
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
            Console.WriteLine("ListItems.exe [-t TABLE] [-r REGION] [-h]");
            Console.WriteLine("");
            Console.WriteLine(" TABLE is optional, and defaults to CustomersOrdersProducts");
            Console.WriteLine(" REGION is optional, and defaults to us-west-2");
            Console.WriteLine(" -h prints this message and quits");
        }

        static async Task<ScanResponse> GetItemsAsync(bool debug, IAmazonDynamoDB client, string table)
        {
            var response = await client.ScanAsync(new ScanRequest {
                TableName = table
            });

            return response;
        }

        static void Main(string[] args)
        {
            bool debug = false;
            string region = "us-west-2";
            string table = "CustomersOrdersProducts";
            

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
                    case "-t":
                        i++;
                        table = args[i];
                        break;
                    default:
                        break;
                }

                i++;
            }

            if (table == "")
            {
                Console.WriteLine("You must supply a non-empty table name (-t TABLE)");
                return;
            }

            DebugPrint(debug, "Debugging enabled\n");

            DebugPrint(debug, "Table  == " + table + "\n");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            var response = GetItemsAsync(debug, client, table);

            Console.WriteLine("Found " + response.Result.Items.Count.ToString() + " items in table " + table + " in region " + region + ":\n");

            StringBuilder output;

            foreach(var item in response.Result.Items)
            {
                output = new StringBuilder();

                foreach(string attr in item.Keys)
                {
                    if (item[attr].S != null)
                    {
                        output.Append(attr + ": " + item[attr].S + ", ");
                    }
                    else if(item[attr].N != null)
                    {
                        output.Append(attr + ": " + item[attr].N.ToString() + ", ");
                    }
                }

                Console.WriteLine(output.ToString());
            }
        }
    }
}
