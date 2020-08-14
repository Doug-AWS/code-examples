using System;
using System.Collections.Generic;
//using System.Collections.Generic;
using System.Text;

namespace TestQuestion
{
    class Program
    {
        static void DebugPrint(bool debug, string s)
        {
            if (debug) Console.WriteLine(s);
        }
        static int GetNumInts(bool debug, string s, int numInts)
        {
            int finalCount;

            DebugPrint(debug, "Got string: " + s + " and #ints: " + numInts.ToString());

            bool parsed = false;
            int i;
            string[] parts = s.Split(" ");

            DebugPrint(debug, "Got " + parts.Length.ToString() + " part(s)");

            if (parts.Length == 0)
            {
                finalCount = numInts;
            }
            else if (parts.Length == 1)
            {
                DebugPrint(debug, "Parsing " + parts[0].ToString());

                parsed = Int32.TryParse(parts[0], out i);
                if (parsed)
                {
                    DebugPrint(debug, "Returning " + (numInts + 1).ToString());
                    finalCount = numInts + 1;
                }
                else
                {
                    DebugPrint(debug, "Returning " + numInts.ToString());
                    finalCount = numInts;
                }
            }
            else
            {
                // Create new array from p1...pn
                StringBuilder newParts = new StringBuilder();
                for (int index = 1; index < parts.Length; index++)
                {
                    newParts.Append(parts[index] + " ");
                }

                // Lop off the last " "
                string newString = newParts.ToString().Trim();

                parsed = Int32.TryParse(parts[0], out i);
                if (parsed)
                {
                    finalCount = GetNumInts(debug, newString, numInts + 1);
                }
                else
                {
                    finalCount = GetNumInts(debug, newString, numInts);
                }
            }

            return finalCount;            
        }

        static int GetAbsValue(int x, int y)
        {
            int diff = x - y;
            return diff < 0 ? -diff : diff;
        }

        static int GetLargestInt(int[] ints)
        {
            if (ints.Length == 0) throw new ArgumentOutOfRangeException("Integer array is empty");

            int largestInt = ints[0];

            foreach(int i in ints)
            {
                if (i > largestInt) largestInt = i;
            }

            return largestInt;
        }
            //[1, 4, 2]) would return 4

        static Dictionary<string, string> GetTags(string service, string [] tagNames)
        {
            Dictionary<string, string> tags = new Dictionary<string, string>();

            foreach(string s in tagNames)
            {
                tags.Add(s, s + service.ToUpper());
            }

            return tags;
        }

        /*
         * This program takes one argument -s SERVICE, where SERVICE is the official short name of an AWS service, such as "Amazon S3".
         * The remaining arguments are tag names.
         */
        static void Main(string[] args)
        {
            if (args.Length < 2)
            {
                Console.WriteLine("You must supply a service name and at least one tag name");
                return;
            }

            string service = "";
            List<string> tagNameList = new List<string>();
            
            for(int i = 0; i < args.Length;)
            {
                if (args[i] == "-s")
                {
                    i++;
                    service = args[i];
                }
                else
                {
                    tagNameList.Add(args[i]);
                }

                i++;
            }

            Dictionary<string, string> tags = GetTags(service, tagNameList.ToArray());

            foreach(KeyValuePair<string, string> kvp in tags)
            {
                Console.WriteLine("Tag " + kvp.Key + " value: " + kvp.Value);
            }

            /*
            string s1 = "abc 123";
            string s2 = "abc 123 xyz 345";
            string s3 = "abc xyz";
            int i1 = GetNumInts(debug, s1, 0); // Should return 1
            int i2 = GetNumInts(debug, s2, 0); // Should return 2
            int i3 = GetNumInts(debug, s3, 0); // Should return 3
            Console.WriteLine("The number of ints (should be 1) in " + s1 + " is " + i1.ToString());
            Console.WriteLine("The number of ints (should be 2) in " + s2 + " is " + i2.ToString());
            Console.WriteLine("The number of ints (should be 0) in " + s3 + " is " + i3.ToString());
            */


        }
    }
}
