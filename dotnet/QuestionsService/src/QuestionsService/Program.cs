using Amazon.CDK;
using System;
using System.Collections.Generic;
using System.Linq;

namespace QuestionsService
{
    sealed class Program
    {
        public static void Main(string[] args)
        {
            var app = new App();
            new QuestionsServiceStack(app, "QuestionsServiceStack");
            app.Synth();
        }
    }
}
