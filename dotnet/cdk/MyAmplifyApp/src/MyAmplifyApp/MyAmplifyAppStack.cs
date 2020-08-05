using Amazon.CDK;
using Amazon.CDK.AWS.Amplify;
using Amazon.CDK.AWS.CodeBuild;
using Amazon.CDK.AWS.SSM;
using System;
using System.Collections.Generic;

namespace MyAmplifyApp
{
    public class MyAmplifyAppStack : Stack
    {
        internal MyAmplifyAppStack(Construct scope, string id, IStackProps props = null) : base(scope, id, props)
        {
            // Get user, repo
            string user = StringParameter.ValueForStringParameter(this, "amplify-user");
            string repo = StringParameter.ValueForStringParameter(this, "amplify-repo");
            SecretValue token = SecretValue.SecretsManager("my-github-token");

            // The code that defines your stack goes here
            Amazon.CDK.AWS.Amplify.App amplifyApp = new Amazon.CDK.AWS.Amplify.App(this, "MyApp", new Amazon.CDK.AWS.Amplify.AppProps
            {
                SourceCodeProvider = new GitHubSourceCodeProvider(new GitHubSourceCodeProviderProps
                {
                    Owner = user,
                    Repository = repo,
                    OauthToken = token
                }),
                BuildSpec = BuildSpec.FromObject(new Dictionary<string, object> { // Alternatively add a `amplify.yml` to the repo
                    { "version", "1.0" },
                    { "frontend", new Dictionary<string, object> {
                        { "Phases", new Dictionary<string, object> {
                            { "PreBuild", new Dictionary<string, object> {
                                { "Commands", new [] { "yarn" } } }
                            },
                            { "Build", new Dictionary<string, object> {
                                { "Commands", new [] { "yarn build" } }
                            }}
                        }},
                        { "Artifacts", new Dictionary<string, object> {
                            { "BaseDirectory", "public" },
                            { "Files", "**/*" }
                        }}
                    }
                }})
            });
        }
    }
}
