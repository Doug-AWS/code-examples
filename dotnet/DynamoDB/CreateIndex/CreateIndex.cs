using System;
using System.Collections.Generic;
using System.Configuration;
using System.Text;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.Model;

namespace CreateIndex
{
    class CreateIndex
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
            Console.WriteLine("CreateIndex.exe -i INDEX-NAME -m MAIN-KEY -s SECONDARY-KEY [-h] [-d]");
            Console.WriteLine("");
            Console.WriteLine("  INDEX-NAME is required");
            Console.WriteLine("  MAIN-KEY is the partition key");
            Console.WriteLine("  SECONDARY-KEY is the sort key");
            Console.WriteLine("");
            Console.WriteLine("  -h prints this message and quits");
            Console.WriteLine("  -d print some extra (debugging) info");
        }

        static async Task<UpdateTableResponse> AddIndexAsync(bool debug, IAmazonDynamoDB client, string table, string indexname, string partitionkey, string sortkey)
        {
            if (null == client)
            {
                throw new ArgumentNullException("client parameter is null");
            }

            if (string.IsNullOrEmpty(table))
            {
                throw new ArgumentNullException("table parameter is null");
            }

            if (string.IsNullOrEmpty(indexname))
            {
                throw new ArgumentNullException("indexname parameter is null");
            }

            if (string.IsNullOrEmpty(partitionkey))
            {
                throw new ArgumentNullException("partitionkey parameter is null");
            }

            if (string.IsNullOrEmpty(sortkey))
            {
                throw new ArgumentNullException("sortkey parameter is null");
            }

            ProvisionedThroughput pt = new ProvisionedThroughput
            {
                ReadCapacityUnits = 10L,
                WriteCapacityUnits = 5L
            };

            KeySchemaElement kse1 = new KeySchemaElement
            {
                AttributeName = partitionkey,
                KeyType = "HASH"
            };

            KeySchemaElement kse2 = new KeySchemaElement
            {
                AttributeName = sortkey,
                KeyType = "RANGE"
            };

            List<KeySchemaElement> kses = new List<KeySchemaElement>
            {
                kse1,
                kse2
            };

            Projection p = new Projection
            {
                ProjectionType = "ALL"
            };

            var newIndex = new CreateGlobalSecondaryIndexAction()
            {
                IndexName = indexname,
                ProvisionedThroughput = pt,
                KeySchema = kses,
                Projection = p
            };

            GlobalSecondaryIndexUpdate update = new GlobalSecondaryIndexUpdate
            {
                Create = newIndex
            };

            List<GlobalSecondaryIndexUpdate> updates = new List<GlobalSecondaryIndexUpdate>
            {
                update
            };

            AttributeDefinition ad1 = new AttributeDefinition
            {
                AttributeName = partitionkey,
                AttributeType = "S"
            };

            AttributeDefinition ad2 = new AttributeDefinition
            {
                AttributeName = sortkey,
                AttributeType = "S"
            };

            UpdateTableRequest request = new UpdateTableRequest
            {
                TableName = table,
                AttributeDefinitions = {
                    ad1, 
                    ad2
                },
                GlobalSecondaryIndexUpdates = updates
            };

            if (debug)
            {
                Console.WriteLine("Update table request:");
                Console.WriteLine(request);
                Console.WriteLine("");
            }

            var response = await client.UpdateTableAsync(request);

            return response;
        }        

        static void Main(string[] args)
        {
            var debug = false;
            var configfile = "../../../../Config/app.config";
            var region = "";
            var table = "";
            var indexname = "";
            var mainkey = "";
            string secondarykey = "";

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
                        indexname = args[i];
                        break;
                    case "-m":
                        i++;
                        mainkey = args[i];
                        break;
                    case "-s":
                        i++;
                        secondarykey = args[i];
                        break;
                    default:
                        break;
                }

                i++;
            }

            bool empty = false;
            StringBuilder sb = new StringBuilder("You must supply a non-empty ");

            if (indexname == "")
            {
                empty = true;
                sb.Append("index name (-i INDEX), ");
            }
            else
            {
                DebugPrint(debug, "Index name: " + indexname);
            }

            if (mainkey == "")
            {
                empty = true;
                sb.Append("mainkey (-m PARTITION-KEY), ");
            }
            else
            {
                DebugPrint(debug, "Main key: " + mainkey);
            }

            if (secondarykey == "")
            {
                empty = true;
                sb.Append("secondary key (-s SORT-KEY), ");
            }
            else
            {
                DebugPrint(debug, "Secondary key: " + secondarykey);
            }

            if (empty)
            {
                Console.WriteLine(sb.ToString());
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

            Task<UpdateTableResponse> response = AddIndexAsync(debug, client, table, indexname, mainkey, secondarykey);

            Console.WriteLine("Task status: " + response.Status);
            Console.WriteLine("Result status: " + response.Result.HttpStatusCode);
        }
    }
}
