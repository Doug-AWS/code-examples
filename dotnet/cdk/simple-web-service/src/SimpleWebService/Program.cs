using Amazon.CDK;
using System;
using System.Collections.Generic;
using System.Linq;

namespace SimpleWebService
{
    sealed class Program
    {
        public static void Main(string[] args)
        {
            var app = new App();
            new SimpleWebServiceStack(app, "SimpleWebServiceStack");
            app.Synth();
        }
    }
}
