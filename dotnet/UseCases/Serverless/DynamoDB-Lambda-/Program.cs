using System;
using System.Collections.Generic;
using System.IO;
using System.Threading.Tasks;

using Newtonsoft.Json;

using Amazon.DynamoDBv2;
//using Amazon.DynamoDBv2.DocumentModel;
using Amazon.DynamoDBv2.Model;
using System.Threading;
using System.Runtime.CompilerServices;
//using System.Diagnostics;
//using System.Linq;

namespace DynamoDB_Lambda_
{
    public class MyQuestion
    {
        public string Question;
        public string Answer1;
        public string Answer2;
        public string Answer3;
        public int Which;
        public string Area;
        public string Level;
        public string Id;

        public MyQuestion(string question, string answer1, string answer2, string answer3, int which, string area, string level, string id)
        {
            Question = question;
            Answer1 = answer1;
            Answer2 = answer2;
            Answer3 = answer3;
            Which = which;
            Area = area;
            Level = level;
            Id = id;
        }
    }

    public class MyConfig
    {
        public string Table;
        public bool Debug;
        public bool Save;

        public MyConfig(string table, bool debug = false, bool save = false)
        {
            Table = table;
            Debug = debug;
            Save = save;
        }
    }
    class Program
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug)
            {
                Console.WriteLine(s);
            }
        }

        static async Task<bool> TableExists(string table)
        {
            bool exists = false;

            AmazonDynamoDBClient client = new AmazonDynamoDBClient();

            try
            {
                var resp = await client.DescribeTableAsync(tableName: table);
                exists = resp.HttpStatusCode == System.Net.HttpStatusCode.OK;
            }
            catch (ResourceNotFoundException)
            {
                exists = false;
            }

            return exists;
        }

        private static async Task<bool> WaitTillTableCreated(bool debug, string table)
        {
            string status = "";
            int sleepTime = 1000; // Initial sleep value of 1 second
            int maxWait = 30000;  // Don't wait more than 30 seconds
            AmazonDynamoDBClient client = new AmazonDynamoDBClient();

            try
            {
                while (status != "ACTIVE")
                {
                    int seconds = sleepTime / 1000;
                    DebugPrint(debug, "Waiting " + seconds + " second(s) for table to be ACTIVE");

                    System.Threading.Thread.Sleep(sleepTime);
                    var resp = await client.DescribeTableAsync(tableName: table);

                    status = resp.Table.TableStatus;
                    sleepTime *= 2;

                    if (sleepTime > maxWait)
                    {
                        DebugPrint(debug, "Creating the table took more than " + maxWait + " seconds");
                        return false;
                    }
                }

                DebugPrint(debug, "Table " + table + " is now ACTIVE");
                return true;
            }
            // Potential eventual-consistency issue.
            catch (ResourceNotFoundException)
            {
                // throw ex;
            }

            return false;
        }

        static async Task<bool> CreateTable(bool debug, string table)
        {
            bool result = false;

            AmazonDynamoDBClient client = new AmazonDynamoDBClient();

            try
            {
                var createResponse = await client.CreateTableAsync(new CreateTableRequest
                {
                    TableName = table,
                    AttributeDefinitions = new List<AttributeDefinition>()
                    {
                        new AttributeDefinition
                        {
                            AttributeName = "id",
                            AttributeType = "S"
                        }
                    },
                    KeySchema = new List<KeySchemaElement>()
                    {
                        new KeySchemaElement
                        {
                            AttributeName = "id",
                            KeyType = "HASH"
                        }
                    },
                    ProvisionedThroughput = new ProvisionedThroughput
                    {
                        ReadCapacityUnits = 10,
                        WriteCapacityUnits = 5
                    }
                });

                Task<bool> waiting = WaitTillTableCreated(debug, table);

                result = waiting.Result;
            }
            catch (ResourceNotFoundException)
            {
                return false;
            }

            return result;
        }

        public static async Task<MyQuestion[]> GetQuestions(bool debug, string table, string areaWanted="all", string levelWanted="all")
        {
            MyQuestion newQuestion;
            List<MyQuestion> questions = new List<MyQuestion>();

            AmazonDynamoDBClient client = new AmazonDynamoDBClient();

            try
            {
                var response = await client.ScanAsync(new ScanRequest
                {
                    TableName = table
                });

                // Which match the area and level
                foreach (Dictionary<string, AttributeValue> item in response.Items)
                {
                    string question = "";
                    string answer1 = "";
                    string answer2 = "";
                    string answer3 = "";
                    int which = 0;
                    string area = "";
                    string level = "";
                    string id = "";

                    foreach (KeyValuePair<string, AttributeValue> kvp in item)
                    {
                        string attributeName = kvp.Key;
                        AttributeValue value = kvp.Value;

                        switch (attributeName)
                        {
                            case "question":
                                question = value.S;
                                break;
                            case "answer1":
                                answer1 = value.S;
                                break;
                            case "answer2":
                                answer2 = value.S;
                                break;
                            case "answer3":
                                answer3 = value.S;
                                break;
                            case "which":
                                which = int.Parse(value.N);
                                break;
                            case "area":
                                area = value.S;
                                break;
                            case "level":
                                level = value.S;
                                break;
                            case "id":
                                id = value.S;
                                break;
                            case "Id":
                                id = value.S;
                                break;
                            default:
                                throw new ArgumentException("Unrecognized attribute name: " + attributeName);
                        }
                    }

                    DebugPrint(debug, "Got question with area: " + area + " and level: " + level);

                    if ((areaWanted == "all" || areaWanted == area) && (levelWanted == "all" || levelWanted == level))
                    {
                        newQuestion = new MyQuestion(question, answer1, answer2, answer3, which, area, level, id);
                        questions.Add(newQuestion);
                    }
                }
            }
            catch (Exception ex)
            {
                throw ex;
            }

            return questions.ToArray();
        }

        public static void ShowData(bool debug, MyQuestion[] questions, string area, string level)
        {
            DebugPrint(debug, "Found " + questions.Length + " questions to display");

            foreach (MyQuestion q in questions)
            {
                if ((area == "all" || area == q.Area) && (level == "all" || level == q.Level))
                {
                    Console.WriteLine("Question: " + q.Question);
                    Console.WriteLine("Answer1:  " + q.Answer1);
                    Console.WriteLine("Answer2   " + q.Answer2);
                    Console.WriteLine("Answer3:  " + q.Answer3);
                    Console.WriteLine("Which:    " + q.Which.ToString());
                    Console.WriteLine("Area:     " + q.Area);
                    Console.WriteLine("Level:    " + q.Level);
                    Console.WriteLine("ID:       " + q.Id);
                    Console.WriteLine("");
                }
            }
        }

        public static void SaveData(bool debug, string folder, string table, MyQuestion[] questions)
        {
            DebugPrint(debug, "Found " + questions.Length + " questions to save");

            // Create file TABLE-DATE.txt for writing
            string archiveName = folder + "\\" + table + "-" + DateTime.Now.ToString("yyyy'-'MM'-'dd") + ".txt";
            // Open the file for editing
            using (StreamWriter outputFile = new StreamWriter(archiveName))
            {
                foreach (MyQuestion q in questions)
                {
                    outputFile.WriteLine("Question: " + q.Question);
                    outputFile.WriteLine("Answer1:  " + q.Answer1);
                    outputFile.WriteLine("Answer2   " + q.Answer2);
                    outputFile.WriteLine("Answer3:  " + q.Answer3);
                    outputFile.WriteLine("Which:    " + q.Which.ToString());
                    outputFile.WriteLine("Area:     " + q.Area);
                    outputFile.WriteLine("Level:    " + q.Level);
                    outputFile.WriteLine("ID:       " + q.Id);
                    outputFile.WriteLine("");
                }
            }

            DebugPrint(debug, "Saved questions to " + archiveName);
        }

        static void Main(string[] args)
        {
            // The full path to the configuration file.
            string configFile = "";

            // The table name. Overrides what's in the config file.
            string tableName = "";

            // Whether to display additional information. Overrides what's in the config file.
            bool debug = false;

            // Whether to save the table information to TABLE-NAME-DATE.txt. Overrides what's in the config file.
            bool save = false;

            // Whether to show questions that match area and level
            bool printData = false;
            string area = "all";
            string level = "all";

            for (int i = 0; i < args.Length; i++)
            {
                switch (args[i])
                {
                    case "-a":
                        i++;
                        area = args[i];
                        break;
                    case "-c":
                        i++;
                        configFile = args[i];
                        break;
                    case "-d":
                        debug = true;
                        break;
                    case "-l":
                        i++;
                        level = args[i];
                        break;
                    case "-p":
                        printData = true;
                        break;
                    case "-s":
                        save = true;
                        break;
                    case "-t":
                        i++;
                        tableName = args[i];
                        break;

                    default:
                        break;
                }
            }

            if (configFile == "")
            {
                Console.WriteLine("You must supply the full path to a configuration file (-c CONFIG-FILE)");
                return;
            }

            MyConfig globalConfiguration;

            DebugPrint(debug, "Getting configuration values from " + configFile);

            using (StreamReader sr = new StreamReader(configFile))
            {
                // Read entire file as a string
                string content = sr.ReadToEnd();
                globalConfiguration = JsonConvert.DeserializeObject<MyConfig>(content);
            }

            // Override if set
            if (tableName != "")
            {
                globalConfiguration.Table = tableName;
            }

            tableName = globalConfiguration.Table;

            if (debug)
            {
                globalConfiguration.Debug = true;
            }

            debug = globalConfiguration.Debug;

            if (save)
            {
                globalConfiguration.Save = true;
            }

            save = globalConfiguration.Save;

            DebugPrint(debug, "Table name: " + tableName);

            // Make sure table exists
            Task<bool> exists = TableExists(tableName);

            if (exists.Result)
            {
                DebugPrint(debug, "The DynamoDB table " + tableName + " exists");

                // Get questions from table
                Task<MyQuestion[]> qs = GetQuestions(debug, tableName);

                // We won't have any results if we just created the table
                if (qs.Result != null)
                {
                    MyQuestion[] questions = qs.Result;

                    if (save)
                    {
                        DebugPrint(debug, "Saving questions");
                        // Get folder where config file resides
                        string folder = Path.GetDirectoryName(configFile);
                        SaveData(debug, folder, tableName, questions);
                    }
                    else
                    {
                        DebugPrint(debug, "NOT saving questions");
                    }

                    if (printData)
                    {
                        DebugPrint(debug, "Showing questions for area: " + area + " and level: " + level);
                        ShowData(debug, questions, area, level);
                    }
                    else
                    {
                        DebugPrint(debug, "NOT printing questions");
                    }
                }
            }
            else
            {
                DebugPrint(debug, "The DynamoDB table " + tableName + " does NOT exist");

                // Create it
                Task<bool> results = CreateTable(debug, tableName);

                if (results.Result)
                {
                    Console.WriteLine("Created table " + tableName);
                }
                else
                {
                    Console.WriteLine("Could NOT create table " + tableName);
                }

                if (save)
                {
                    Console.WriteLine("Cannot save new table");
                }

                Console.WriteLine("Press enter to finish");
                string response = Console.ReadLine();
            }
        }
    }
}
