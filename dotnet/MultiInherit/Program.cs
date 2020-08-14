using System;

namespace MultiInherit
{
    class ClassOne
    {
        public ClassOne()
        {
            Console.WriteLine("In ClassOne constructor");
        }

        public void SayIt(string s)
        {
            Console.WriteLine(s + ": ClassOne");
        }
    }

    class ClassTwo: ClassOne
    {
        public ClassTwo()
        {
            Console.WriteLine("In ClassTwo constructor");
        }

        public void SayIt(string s)
        {
            Console.WriteLine(s + ": ClassTwo");
        }
    }

    class Program: ClassTwo
    {
        public Program()
        {
            Console.WriteLine("In program constructor");
        }

        public void SayIt(string s)
        {
            Console.WriteLine(s + ": Program");
        }

        static void Main(string[] args)
        {
            Program p = new Program();
            Console.WriteLine("In Main");

            p.SayIt("Inside");
        }
    }
}
