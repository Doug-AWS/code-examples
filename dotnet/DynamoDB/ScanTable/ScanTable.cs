using System;
using System.Collections.Generic;
using System.Configuration;
using System.Globalization;
using System.Text;
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
            Console.WriteLine("ScanTable.exe [-l(ow) | -r(ange) | -p(roducts)] [-m MINIMUM] [-s START] [-e END] [-i ID]  [-d] [-h]");
            Console.WriteLine("");
            Console.WriteLine("  low gets any product with a quantity below MINIMUM");
            Console.WriteLine("  range gets any orders between START and END");
            Console.WriteLine("  products gets any orders for products with product ID ID");
            Console.WriteLine("  the default is low, but if you set more than one, the last one wins");
            Console.WriteLine("");
            Console.WriteLine("  MINIMUM-QUANTITY must be > 0; default is 100");
            Console.WriteLine("  START must be a date in the format: yyyy-MM-dd HH:mm:ss; default is 2020-05-04 05:00:00");
            Console.WriteLine("  END must be a date in the format: yyyy-MM-dd HH:mm:ss; default is 2020-08-13 09:00:00");
            Console.WriteLine("  ID must be an integer; default is 3");
            Console.WriteLine("");
            Console.WriteLine(" -h prints this message and quits");
            Console.WriteLine(" -d prints extra (debugging) info");
        }

        // Get the orders made in range from start to end
        // DynamoDB equivalent of:
        //   select * from Orders where Order_Date between '2020-05-04 05:00:00' and '2020-08-13 09:00:00'
        static async Task<ScanResponse> GetOrdersInDateRangeAsync(IAmazonDynamoDB client, string table, string start, string end)
        {
            // Convert start and end strings to longs
            var StartDateTime = DateTime.ParseExact(start, "yyyy-MM-dd HH:mm:ss", CultureInfo.InvariantCulture);
            var EndDateTime = DateTime.ParseExact(end, "yyyy-MM-dd HH:mm:ss", CultureInfo.InvariantCulture);

            TimeSpan startTimeSpan = StartDateTime - new DateTime(1970, 1, 1, 0, 0, 0);
            TimeSpan endTimeSpan = EndDateTime - new DateTime(1970, 1, 1, 0, 0, 0);

            var begin = (long)startTimeSpan.TotalSeconds;
            var finish = (long)endTimeSpan.TotalSeconds;

            var response = await client.ScanAsync(new ScanRequest
            {
                TableName = table,
                ExpressionAttributeValues = new Dictionary<string, AttributeValue> {
                    {":startval", new AttributeValue { N = begin.ToString() } },
                    {":endval", new AttributeValue { N = finish.ToString()} }
                },
                FilterExpression = "Order_Date > :startval AND Order_Date < :endval",
                ProjectionExpression = "Order_ID, Order_Customer, Order_Product, Order_Date, Order_Status"
            });

            return response;
        }

        // Get the orders for product with ID productId
        // DynamoDB equivalent of:
        //   select* from Orders where Order_Product = '3'
        static async Task<ScanResponse> GetProductOrdersAsync(bool debug, IAmazonDynamoDB client, string table, string productId)
        {
            DebugPrint(debug, "Retrieving orders where the product ordered had ID " + productId);

            var response = await client.ScanAsync(new ScanRequest
            {
                TableName = table,
                ExpressionAttributeValues = new Dictionary<string, AttributeValue> {
                    {":val", new AttributeValue { N = productId }}
                },
                FilterExpression = "Order_Product = :val",
                ProjectionExpression = "Order_ID,Order_Customer,Order_Product,Order_Date,Order_Status"
            });

            return response;
        }

        // Get the products with less than volume items in the warehouse
        // DynamoDB equivalent of:
        //   select* from Products where Product_Quantity < '100'
        static async Task<ScanResponse> GetLowStockAsync(bool debug, IAmazonDynamoDB client, string table, string volume)
        {
            DebugPrint(debug, "Retrieving all products with fewer than " + volume + " items in stock");

            var response = await client.ScanAsync(new ScanRequest
            {
                TableName = table,
                ExpressionAttributeValues = new Dictionary<string, AttributeValue> {
                    {":val", new AttributeValue { N = volume }}
                },
                FilterExpression = "Product_Quantity < :val",
                ProjectionExpression = "Product_ID, Product_Description, Product_Quantity, Product_Cost"
            });

            return response;
        }

        static void Main(string[] args)
        {
            var debug = false;
            var configfile = "app.config";
            var region = "";
            var table = "";
            var query = "products";
            var start = "2020-05-04 05:00:00";
            var end = "2020-08-13 09:00:00";
            var id = "3";
            string minimum = "100";

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
                    case "-l":
                        query = "low";
                        break;
                    case "-r":
                        query = "range";
                        break;
                    case "-p":
                        query = "products";
                        break;
                    case "-s":
                        i++;
                        start = args[i];
                        break;
                    case "-e":
                        i++;
                        end = args[i];
                        break;
                    case "-m":
                        i++;
                        minimum = args[i];
                        break;
                    default:
                        break;
                }

                i++;
            }

            // Get default region and table from config file
            var efm = new ExeConfigurationFileMap
            {
                ExeConfigFilename = configfile
            };

            Configuration configuration = ConfigurationManager.OpenMappedExeConfiguration(efm, ConfigurationUserLevel.None);

            if (configuration.HasFile)
            {
                AppSettingsSection appSettings = configuration.AppSettings;
                region = appSettings.Settings["Region"].Value;
                table = appSettings.Settings["Table"].Value;

                if ((region == "") || (table == ""))
                {
                    Console.WriteLine("You must specify Region and Table values in " + configfile);                    
                    return;
                }
            }
            else
            {
                Console.WriteLine("Could not find " + configfile);
                return;
            }

            DebugPrint(debug, "Debugging enabled");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<ScanResponse> response = null;

            switch (query)
            {
                case "low":
                    // Make sure we have a minimum value
                    if (minimum == "")
                    {
                        Console.WriteLine("You must supply a non-empty MINIMUM value (-m MINIMUM, ");
                        Usage();
                        return;
                    }

                    int number;

                    try
                    {
                        number = int.Parse(minimum);

                        if (number < 1)
                        {
                            Console.WriteLine("The minimum quantity must be > 0");
                            Usage();
                            return;
                        }
                    }
                    catch (FormatException)
                    {
                        Console.WriteLine(minimum + " is not an integer");
                        Usage();
                        return;
                    }

                    response = GetLowStockAsync(debug, client, table, minimum);
                    break;
                case "range":
                    // Make sure we have start and end values,
                    // and that they are proper dates
                    if ((start == "") || (end == ""))
                    {
                        Console.WriteLine("The range option requires START and END flag values");
                        Usage();
                        return;
                    }                    

                   response = GetOrdersInDateRangeAsync(client, table, start, end);

                    break;
                case "products":
                    // Make sure we have a product ID
                    if (id == "")
                    {
                        Console.WriteLine("The products option requires an ID flag value");
                        Usage();
                        return;
                    }

                    int val;

                    try
                    {
                        val = int.Parse(id);

                        if (val < 1)
                        {
                            Console.WriteLine("The product ID must be > 0");
                            Usage();
                            return;
                        }
                    }
                    catch (FormatException)
                    {
                        Console.WriteLine(id + " is not an integer");
                        Usage();
                        return;
                    }

                    response = GetProductOrdersAsync(debug, client, table, id);

                    break;
                default:
                    break;
            }

            // To adjust date/time value
            var epoch = new DateTime(1970, 1, 1, 0, 0, 0, DateTimeKind.Utc);

            DebugPrint(debug, "Today's # ticks: " + DateTime.Now.Ticks.ToString());

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
                        // If the attribute contains the string "date", process it differently
                        if (attr.ToLower().Contains("date"))
                        {
                            long span = long.Parse(item[attr].N);                            
                            DateTime theDate = epoch.AddSeconds(span);        
                            
                            Console.WriteLine(attr + ": " + theDate.ToLongDateString());
                        }
                        else
                        {
                            Console.WriteLine(attr + ": " + item[attr].N.ToString());
                        }
                    }
                }

                Console.WriteLine("");
            }
        }
    }
}
