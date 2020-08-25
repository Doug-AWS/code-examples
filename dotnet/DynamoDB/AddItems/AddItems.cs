using System;
using System.Collections.Generic;
using System.Globalization;
using System.IO;
using System.Net.NetworkInformation;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.DocumentModel;

using Newtonsoft.Json;

namespace DynamoDBCRUD
{    
    class AddItems
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }
                
        public static async Task<int> AddFromCSVAsync(bool debug, IAmazonDynamoDB client, string table, string filename, int index)
        {
            string line;
            Table theTable = Table.LoadTable(client, table);
            Document item = new Document();

            // filename is the name of the csv file that contains customer data
            // in lines 2...N
            // Column1,...,ColumnN
            // Read the file and display it line by line.  
            System.IO.StreamReader file =
                new System.IO.StreamReader(filename);

            // Get column names from the first line
            string [] headers = file.ReadLine().Split(",");
            int numcolumns = headers.Length;

            int lineNum = 2;

            while ((line = file.ReadLine()) != null)
            {
                // Split line into columns
                string[] parts = line.Split(',');

                // if we don't have the right number of parts, something's wrong
                if (parts.Length != numcolumns)
                {                    
                    Console.WriteLine("Did not have " + numcolumns.ToString() + " columns in line " + lineNum.ToString() + " of file " + filename);
                    return 0;
                }

                item["ID"] = index.ToString();

                DebugPrint(debug, "Adding item with index " + index.ToString() + " to table");

                index++;

                for (int i = 0; i < numcolumns; i++)
                {
                    // if the header contains the word "date", store the value as a long (number)
                    if (headers[i].ToLower().Contains("date"))
                    {
                        // The datetime format is:
                        // YYYY-MM-DD HH:MM:SS
                        DateTime MyDateTime = DateTime.ParseExact(parts[i], "yyyy-MM-dd HH:mm:ss", CultureInfo.InvariantCulture);

                        TimeSpan timeSpan = MyDateTime - new DateTime(1970, 1, 1, 0, 0, 0);

                        item[headers[i]] = (long)timeSpan.TotalSeconds;
                    }
                    else
                    {
                        // If it's a number, store it as such
                        try
                        {
                            int v = int.Parse(parts[i]);
                            item[headers[i]] = v;
                        }
                        catch
                        {
                            item[headers[i]] = parts[i];
                        }
                    }
                }

                await theTable.PutItemAsync(item);

                lineNum++;
            }

            file.Close();

            return index;
        }

        static void Main(string[] args)
        {
            int index = 0;
            bool debug = false;
            string table = "";
            string region = "us-west-2";
            string customers = "";
            string orders = "";
            string products = "";

            int i = 0;

            while (i < args.Length)
            {
                switch (args[i])
                {
                    case "-c":
                        i++;
                        customers = args[i];
                        break;
                    case "-d":
                        debug = true;
                        break;
                    case "-o":
                        i++;
                        orders = args[i];
                        break;
                    case "-p":
                        i++;
                        products = args[i];
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

            if ((region == "") || (table == "") || (customers == "") || (orders == "") || (products == ""))
            {
                Console.WriteLine("You must include a non-empty region (-r REGION), and customers (-c CUSTOMERS-FILE.csv), orders (-o ORDERS-FILE.csv), and products (-p PRODUCTS-FILE.csv) files");
                return;
            }

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            DebugPrint(debug, "Adding customers from " + customers);
            Task<int> result = AddFromCSVAsync(debug, client, table, customers, index);

            index = result.Result;

            if (index == 0)
            {
                return;
            }

            DebugPrint(debug, "Adding orders from " + orders);

            result = AddFromCSVAsync(debug, client, table, orders, index);

            index = result.Result;

            if (index == 0)
            {
                return;
            }

            DebugPrint(debug, "Adding products from " + products);

            result = AddFromCSVAsync(debug, client, table, products, index);

            index = result.Result;

            if (index == 0)
            {
                return;
            }

            Console.WriteLine("Done");
        }
    }
}
