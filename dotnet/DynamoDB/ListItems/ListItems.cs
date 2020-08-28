using System;
using System.Configuration;
using System.Text;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

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
            Console.WriteLine("ListItems.exe [-r REGION] [-h]");
            Console.WriteLine("");
            Console.WriteLine(" -h prints this message and quits");
        }

        static async Task<ScanResponse> GetItemsAsync(IAmazonDynamoDB client, string table)
        {
            var response = await client.ScanAsync(new ScanRequest {
                TableName = table
            });

            return response;
        }        

        static void Main(string[] args)
        {
            bool debug = false;
            var configfile = "app.config";
            var region = "";
            var table = "";

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
            }
            else
            {
                Console.WriteLine("Could not find " + configfile);
                return;
            }

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

            var empty = false;
            var sb = new StringBuilder("You must supply a non-empty ");

            if (table == "")
            {
                empty = true;
                sb.Append("table name (-t TABLE), ");
            }
            else
            {
                DebugPrint(debug, "Table: " + table + "\n");
            }

            if (region == "")
            {
                empty = true;
                sb.Append("region -r (REGION)");
            }
            else
            {
                DebugPrint(debug, "Region: " + region);
            }

            if (empty)
            {
                Console.WriteLine(sb.ToString());
                return;
            }

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            var response = GetItemsAsync(client, table);

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
