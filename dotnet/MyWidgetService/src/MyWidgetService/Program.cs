using Amazon.CDK;
using System;
using System.Collections.Generic;
using System.Linq;

namespace MyWidgetService
{
    sealed class Program
    {
        public static void Main(string[] args)
        {
            var app = new App();
            new MyWidgetServiceStack(app, "MyWidgetServiceStack");
            app.Synth();
        }
    }
}
