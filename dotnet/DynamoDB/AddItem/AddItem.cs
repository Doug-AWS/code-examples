using System;
using System.Collections.Generic;
using System.Globalization;
using System.Reflection;
using System.Threading;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.DocumentModel;
using Amazon.DynamoDBv2.Model;

namespace DynamoDBCRUD
{
    class AddItem
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }

        static async Task<bool> AddItemAsync(bool debug, IAmazonDynamoDB client, string table, string keystring, string valuestring)
        {
            // Get individual keys and values
            string[] keys = keystring.Split(",");
            string[] values = valuestring.Split(",");

            if (keys.Length != values.Length)
            {
                Console.WriteLine("Unmatched number of keys and values");
                return false;
            }

            var theTable = Table.LoadTable(client, table);
            var item = new Document();

            for(int i = 0; i < keys.Length; i++)
            {
                // if the header contains the word "date", store the value as a long (number)
                if (keys[i].ToLower().Contains("date"))
                {
                    // The datetime format is:
                    // YYYY-MM-DD HH:MM:SS
                    DateTime MyDateTime = DateTime.ParseExact(values[i], "yyyy-MM-dd HH:mm:ss", CultureInfo.InvariantCulture);

                    TimeSpan timeSpan = MyDateTime - new DateTime(1970, 1, 1, 0, 0, 0);

                    item[keys[i]] = (long)timeSpan.TotalSeconds;
                }
                else
                {
                    // If it's a number, store it as such
                    try
                    {
                        int v = int.Parse(values[i]);
                        item[keys[i]] = v;
                    }
                    catch
                    {
                        item[keys[i]] = values[i];
                    }
                }
            }

            await theTable.PutItemAsync(item);

            return true;
        }

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("AddItem.exe -k KEYS -v VALUES [-t TABLE] [-r REGION] [-h]");
            Console.WriteLine("");
            Console.WriteLine("Both KEYS and VALUES are required");
            Console.WriteLine("Both should be a comma-separated list, and must have the same number of values");
            Console.WriteLine(" TABLE is optional, and defaults to CustomersOrdersProducts");
            Console.WriteLine(" REGION is optional, and defaults to us-west-2");
            Console.WriteLine(" -h prints this message and quits");
        }

        static void Main(string[] args)
        {
            bool debug = false;
            string region = "us-west-2";
            string table = "CustomersOrdersProducts";
            string keys = "";
            string values = "";

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
                    case "-k":
                        i++;
                        keys = args[i];
                        break;
                    case "-v":
                        i++;
                        values = args[i];
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

            if ((table == "") || (keys == "") || (values == ""))
            {
                Console.WriteLine("You must supply a non-empty table name (-t TABLE), comma-separate list of keys (-k KEYS) and comma-separated list of values (-v VALUES)");
                return;
            }

            DebugPrint(debug, "Debugging enabled\n");

            DebugPrint(debug, "Table  == " + table + "\n");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

           Task<bool> response = AddItemAsync(debug, client, table, keys, values);

            Console.WriteLine("Added item to " + table + " in region " + region);
        }
    }
}
