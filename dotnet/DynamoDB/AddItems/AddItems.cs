using System;
using System.Configuration;
using System.Globalization;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.DocumentModel;

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

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("AddItems.exe [-d] [-h]");
            Console.WriteLine("");
            Console.WriteLine("  -d prints extra (debugging) info");
            Console.WriteLine("  -h prints this message and quits");
        }

        public static async Task<int> AddFromCSVAsync(bool debug, IAmazonDynamoDB client, string table, string filename, int index)
        {
            DebugPrint(debug, "Loading data from file " + filename + " into table " + table);

            var theTable = Table.LoadTable(client, table);
            var item = new Document();

            // filename is the name of the csv file that contains customer data
            // Column1,...,ColumnN
            // in lines 2...N
            // Read the file and display it line by line.  
            System.IO.StreamReader file =
                new System.IO.StreamReader(filename);

            // Get column names from the first line
            string firstline = file.ReadLine();
            DebugPrint(debug, "Columns:");
            DebugPrint(debug, firstline);

            string [] headers = firstline.Split(",");
            int numcolumns = headers.Length;

            var lineNum = 2;
            string line;

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
                        DebugPrint(debug, "Adding date " + parts[i] + " for column " + headers[i] + " to table");
                    }
                    else
                    {
                        // If it's a number, store it as such
                        try
                        {
                            int v = int.Parse(parts[i]);
                            item[headers[i]] = v;
                            DebugPrint(debug, "Adding number " + parts[i] + " for column " + headers[i] + " to table");
                        }
                        catch
                        {
                            item[headers[i]] = parts[i];
                            DebugPrint(debug, "Adding string " + parts[i] + " for column " + headers[i] + " to table");
                        }
                    }
                }

                await theTable.PutItemAsync(item);

                lineNum++;
            }

            file.Close();

            DebugPrint(debug, "Parsed " + lineNum.ToString() + " lines from source file");

            return index;
        }

        static void Main(string[] args)
        {
            bool debug = false;
            var configfile = "app.config";
            var region = "";
            var table = "";
            string customers = "";
            string orders = "";
            string products = "";

            int i = 0;

            while (i < args.Length)
            {
                switch (args[i])
                {
                    case "-d":
                        debug = true;
                        break;
                   case "-h":
                        Usage();
                        return;
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
                customers = appSettings.Settings["Customers"].Value;
                orders = appSettings.Settings["Orders"].Value;
                products = appSettings.Settings["Products"].Value;

                if ((region == "") || (table == "") || (customers == "") || (orders == "") || (products == ""))
                {
                    Console.WriteLine("You must specify Region, Table, Customers, Orders, and Products values in " + configfile);
                    return;
                }
            }
            else
            {
                Console.WriteLine("Could not find " + configfile);
                return;
            }

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            DebugPrint(debug, "Adding customers from " + customers);
            var index = 0;

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
