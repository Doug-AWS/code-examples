using Amazon.CDK;
using System;
using System.Collections.Generic;
using System.Linq;

namespace MyAmplifyApp
{
    sealed class Program
    {
        public static void Main(string[] args)
        {
            var app = new App();
            new MyAmplifyAppStack(app, "MyAmplifyAppStack");
            app.Synth();
        }
    }
}
