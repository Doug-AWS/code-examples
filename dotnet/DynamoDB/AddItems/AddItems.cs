using System;
using System.Collections.Generic;
using System.IO;
using System.Net.NetworkInformation;
using System.Threading.Tasks;

using Amazon;
using Amazon.DynamoDBv2;
using Amazon.DynamoDBv2.DocumentModel;

using Newtonsoft.Json;

namespace DynamoDBCRUD
{
    public class Song
    {
        public string Artist { get; set; }
        public string SongTitle { get; set; }
    }

    public class Songs
    {
        public List<Song> NewSongs { get; set; }
    }

    class AddItems
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
        static Songs GetItems(bool debug, string fileName)
        {
            string jsonString = File.ReadAllText(fileName);

            DebugPrint(debug, "Got JSON:");
            DebugPrint(debug, jsonString);
            DebugPrint(debug, "");

            var songlist = JsonConvert.DeserializeObject<Songs>(jsonString);

            return songlist;        
        }

        public static async void AddItemsAsync(bool debug, IAmazonDynamoDB client, string table, Songs songs)
        {
            if (debug)
            {
                Console.WriteLine("New songs:");

                foreach (Song song in songs.NewSongs)
                {
                    Console.WriteLine("\"" + song.SongTitle + "\" from \"" + song.Artist + "\"");
                }
            }

            var theTable = Table.LoadTable(client, table);
            var item = new Document();

            foreach (Song song in songs.NewSongs)
            {
                item["Artist"] = song.Artist;
                item["SongTitle"] = song.SongTitle;

                await theTable.PutItemAsync(item);
            }
        }

        static void Main(string[] args)
        {
            bool debug = false;
            string table = "Music";
            string region = "us-west-2";
            string fileName = "";

            int i = 0;

            while (i < args.Length)
            {
                switch (args[i])
                {
                    case "-d":
                        debug = true;
                        break;
                    case "-f":
                        i++;
                        fileName = args[i];
                        break;
                    default:
                        break;
                }

                i++;
            }

            DebugPrint(debug, "Adding songs from " + fileName);

            var newRegion = RegionEndpoint.GetBySystemName(region);
            IAmazonDynamoDB client = new AmazonDynamoDBClient(newRegion);

            Songs songs = GetItems(debug, fileName);

            DebugPrint(debug, "Got " + songs.NewSongs.Count.ToString() + " songs from " + fileName);

            AddItemsAsync(debug, client, table, songs);

            Console.WriteLine("Done");
        }
    }
}
