using Amazon.CDK;
using System;
using System.Collections.Generic;
using System.Linq;

namespace MyCognitoUserPool
{
    sealed class Program
    {
        public static void Main(string[] args)
        {
            var app = new App();
            new MyCognitoUserPoolStack(app, "MyCognitoUserPoolStack");
            app.Synth();
        }
    }
}
