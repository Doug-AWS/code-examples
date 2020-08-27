using System;
using System.Collections.Generic;
using System.Configuration;
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

        static async Task<ListTablesResponse> ShowTablesAsync(IAmazonDynamoDB client)
        {
            var response = await client.ListTablesAsync(new ListTablesRequest { });

            return response;
        }

        
        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("ListTables.exe [-h]");
            Console.WriteLine("");
            Console.WriteLine(" -h prints this message and quits");
        }

        static void Main(string[] args)
        {
            var debug = false;
            var region = "";
            var configfile = "../../../../Config/app.config";

            // Get default region from config file
            var efm = new ExeConfigurationFileMap {
                ExeConfigFilename = configfile 
            };
            
            Configuration configuration = ConfigurationManager.OpenMappedExeConfiguration(efm, ConfigurationUserLevel.None);
            
            if (configuration.HasFile)
            {
                AppSettingsSection appSettings = configuration.AppSettings;
                region = appSettings.Settings["Region"].Value;

                if (region == "")
                {
                    Console.WriteLine("You must set a Region value in " + configfile);
                    return;
                }
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
                    default:
                        break;
                }

                i++;
            }

            DebugPrint(debug, "Debugging enabled");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Task<ListTablesResponse> response = ShowTablesAsync(client);

            Console.WriteLine("Found " + response.Result.TableNames.Count.ToString() + " tables in " + region + " region:");

            foreach (var table in response.Result.TableNames)
            {
                Console.WriteLine("  " + table);
            }
        }
    }
}
