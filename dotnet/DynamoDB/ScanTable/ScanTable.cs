using System;
using System.Collections.Generic;
using System.Configuration;
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
            Console.WriteLine("ScanTable.exe -m MINIMUM-QUANTITY [-d] [-h]");
            Console.WriteLine("");
            Console.WriteLine("  MINIMUM-QUANTITY is required, and must be > 0");
            Console.WriteLine(" -h prints this message and quits");
            Console.WriteLine(" -d prints extra (debugging) info");
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
            var debug = false;
            var configfile = "../../../../Config/app.config";
            var region = "";
            var table = "";
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
                    default:
                        break;
                }

                i++;
            }

            if (minimum == "")
            {
                Console.WriteLine("You must supply a non-empty minimum quantity (-m MINIMUM-QUANTITY, ");            
                return;
            }

            int number;

            try
            {
                number = int.Parse(minimum);

                if (number < 1)
                {
                    Console.WriteLine("The minimum quantity must be > 0");
                    return;
                }
            }
            catch (FormatException)
            {
                Console.WriteLine(minimum + " is not an integer");
                return;
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
