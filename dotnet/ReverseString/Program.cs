using System;
using System.Runtime.InteropServices.ComTypes;
using System.Text;

namespace ReverseString
{
    class Program
    {
        public static string FizzBang(int i)
        {
            if (i % 3 == 0 && i % 5 == 0) return "FizzBang";
            if (i % 5 == 0) return "Bang";    
            if (i % 3 == 0) return "Fizz";

            return i.ToString();
        }

        public static string FizzBang2(int i)
        {
            if (i % 3 == 0)
            {
                if (i % 5 == 0)
                {
                    return "FizzBang";
                }
                else
                {
                    return "Fizz";
                }
            }
            else if (i % 5 == 0)
            {
                return "Bang";
            }

            return i.ToString();
        }

        static void Main(string[] args)
        {
            for (int i = 0; i < 100; i++)
            {
                string s = FizzBang(i);
                string s2 = FizzBang2(i);

                Console.WriteLine(s);
                Console.WriteLine(s2);
            }
        }
    }
}
