using System;
using System.Collections.Generic;
using System.IO;
using System.Threading.Tasks;

using Newtonsoft.Json;

using Amazon.DynamoDBv2;
//using Amazon.DynamoDBv2.DocumentModel;
using Amazon.DynamoDBv2.Model;
//using System.Diagnostics;
//using System.Linq;

namespace DynamoDB_Lambda_
{
    public class MyQuestion
    {
        public string question;
        public string answer;
        public string area;
        public string level;
        public string id;

        public MyQuestion(string question, string answer, string area, string level, string id)
        {
            this.question = question;
            this.answer = answer;
            this.area = area;
            this.level = level;
            this.id = id;
        }
    }

    public class MyConfig
    {
        public string Table;
        public bool Debug;
        public bool Archive;

        public MyConfig(string table, bool debug = false, bool archive=false)
        {
            Table = table;
            Debug = debug;
            Archive = archive;
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

        private static async void WaitTillTableCreated(bool debug, string table)
        {
            string status = "";
            int sleepTime = 1000; // Initial sleep value of 1 second
            int maxWait = 30000;  // Don't wait more than 30 seconds
            AmazonDynamoDBClient client = new AmazonDynamoDBClient();

            try
            {
                while (status != "ACTIVE")
                {
                    System.Threading.Thread.Sleep(sleepTime);
                    var resp = await client.DescribeTableAsync(tableName: table);

                    status = resp.Table.TableStatus;
                    sleepTime *= 2;

                    if (sleepTime > maxWait)
                    {
                        throw new TimeoutException("Creating the table took more than " + maxWait + " seconds");
                    }
                }
            }
            // Potential eventual-consistency issue.
            catch (ResourceNotFoundException ex)
            {
                throw ex;
            }
        }

        
    static async void CreateTable(bool debug, string table)
        {
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
                                      AttributeName = "Id",
                                      AttributeType = "S"
                                  }
                              },
                    KeySchema = new List<KeySchemaElement>()
                              {
                                  new KeySchemaElement
                                  {
                                      AttributeName = "Id",
                                      KeyType = "HASH"
                                  }
                              },
                    ProvisionedThroughput = new ProvisionedThroughput
                    {
                        ReadCapacityUnits = 10,
                        WriteCapacityUnits = 5
                    }
                }); ;

                WaitTillTableCreated(debug, table);

                DebugPrint(debug, "Created table " + table);
            }
            catch (ResourceNotFoundException ex)
            {
               throw ex;
            }
        }

        public static async Task<MyQuestion[]> GetQuestions(bool debug, string table)
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
                    string answer = "";
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
                            case "answer":
                                answer = value.S;
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

                    newQuestion = new MyQuestion(question, answer, area, level, id);
                    questions.Add(newQuestion);
                }
            }
            catch (Exception ex)
            {
                throw ex;
            }

            return questions.ToArray();
        }

        public static void ArchiveData(bool debug, string folder, string table, MyQuestion[] questions)
        {
            DebugPrint(debug, "Found " + questions.Length + " questions to archive");

            // Create file TABLE-DATE.txt for writing
            string archiveName = folder + "\\" + table + "-" + DateTime.Now.ToString("yyyy'-'MM'-'dd") + ".txt";
            // Open the file for editing
            using (StreamWriter outputFile = new StreamWriter(archiveName))
            {
                foreach (MyQuestion q in questions)
                {
                    outputFile.WriteLine("Question: " + q.question);
                    outputFile.WriteLine("Answer: " + q.answer);
                    outputFile.WriteLine("Area: " + q.area);
                    outputFile.WriteLine("Level: " + q.level);
                    outputFile.WriteLine("ID: " + q.id);
                    outputFile.WriteLine("");
                }
            }

            DebugPrint(debug, "Archived questions to " + archiveName);
        }

        static void Main(string[] args)
        {
            // The full path to the configuration file.
            string configFile = "";

            // The table name. Overrides what's in the config file.
            string tableName = "";

            // Whether to display additional information. Overrides what's in the config file.
            bool debug = false;

            // Whether to archive the table information to TABLE-NAME-DATE.txt. Overrides what's in the config file.
            bool archive = false;

            for(int i = 0; i < args.Length; i++)
            {
                switch (args[i])
                    {
                    case "-a":
                        archive = true;
                        break;
                    case "-c":
                        i++;
                        configFile = args[i];
                        break;
                    case "-d":
                        debug = true;
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

            if (archive)
            {
                globalConfiguration.Archive = true;
            }

            archive = globalConfiguration.Archive;

            DebugPrint(debug, "Table name: " + tableName);

            // Make sure table exists
            Task<bool> exists = TableExists(tableName);

            if (exists.Result)
            {
                DebugPrint(debug, "The DynamoDB table " + tableName + " exists");
            } else
            {
                DebugPrint(debug, "The DynamoDB table " + tableName + " does NOT exist");

                // Create it
                CreateTable(debug, globalConfiguration.Table);
            }

            // Get questions from table
            Task<MyQuestion[]> qs = GetQuestions(debug, tableName);

            MyQuestion[] questions = qs.Result;

            if (archive)
            {
                // Get folder where config file resides
                string folder = Path.GetDirectoryName(configFile);
                ArchiveData(debug, folder, tableName, questions);
            }

            Console.WriteLine("Press any key to finish");
            string response = Console.ReadLine();
        }
    }
}
