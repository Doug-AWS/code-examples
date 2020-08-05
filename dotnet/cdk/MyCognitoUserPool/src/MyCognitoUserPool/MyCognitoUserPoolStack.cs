using Amazon.CDK;
using Amazon.CDK.AWS.Cognito;
using System;

namespace MyCognitoUserPool
{
    public class MyCognitoUserPoolStack : Stack
    {
        internal MyCognitoUserPoolStack(Construct scope, string id, IStackProps props = null) : base(scope, id, props)
        {
            // The code that defines your stack goes here
            var userpool = new UserPool(this, "myuserpool", new UserPoolProps {
                SignInCaseSensitive = false, // So user can sign in as username, Username, etc.
                SelfSignUpEnabled = true,
                UserPoolName = "MyUserPool",
                UserVerification = new UserVerificationConfig {
                    EmailSubject = "Verify your email for our awesome app!",
                    EmailBody = "Hello {username}, Thanks for signing up to our awesome app! Your verification code is {####}",
                    EmailStyle =  VerificationEmailStyle.CODE,
                    SmsMessage = "Hello {username}, Thanks for signing up to our awesome app! Your verification code is {####}"
                },
                SignInAliases = new SignInAliases {
                    Username = true,
                    Email = true
                }
            });

            userpool.AddDomain("CognitoDomain", new UserPoolDomainProps { // UserPoolDomainProps implements IUserPoolDomainOptions {
                CognitoDomain = new CognitoDomainOptions {                
                    DomainPrefix = "my-awesome-app"
                }
            });

            userpool.AddClient("MyUserPoolClient", new UserPoolClientProps {

            });
        }
    }
}
