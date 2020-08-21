using System;
using System.Collections.Generic;
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

        static async Task<Document> AddItemAsync(bool debug, IAmazonDynamoDB client, string table, string artist, string title)
        {
            var theTable = Table.LoadTable(client, table);
            var item = new Document();

            item["Artist"] = artist;
            item["SongTitle"] = title;

            var response = await theTable.PutItemAsync(item);

            return response;
        }

        static void Usage()
        {
            Console.WriteLine("Usage:");
            Console.WriteLine("AddItem.exe -a ARTIST -s SONG-TITLE [-t TABLE] [-r REGION] [-h]");
            Console.WriteLine("");
            Console.WriteLine("Both ARTIST and SONG-TITLE are required");
            Console.WriteLine(" TABLE is optional, and defaults to Music");
            Console.WriteLine(" REGION is optional, and defaults to us-west-2");
            Console.WriteLine(" -h prints this message and quits");
        }

        static void Main(string[] args)
        {
            bool debug = false;
            string region = "us-west-2";
            string table = "Music";
            string artist = "";
            string title = "";

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
                    case "-a":
                        i++;
                        artist = args[i];
                        break;
                    case "-s":
                        i++;
                        title = args[i];
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

            if ((table == "") || (artist == "") || (title == ""))
            {
                Console.WriteLine("You must supply a non-empty table name (-t TABLE), artist (-a ARTIST) and song title (-s SONG-TITLE)");
                return;
            }

            DebugPrint(debug, "Debugging enabled\n");

            DebugPrint(debug, "Table  == " + table + "\n");

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

           Task<Document> response = AddItemAsync(debug, client, table, artist, title);

            Console.WriteLine("Added artist \"" + artist + "\" with song title \"" + title + "\" to table " + table + " in region " + region);
        }
    }
}
