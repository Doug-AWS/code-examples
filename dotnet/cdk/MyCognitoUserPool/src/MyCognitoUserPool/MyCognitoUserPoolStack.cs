using Amazon.CDK;
using Amazon.CDK.AWS.Cognito;

namespace MyCognitoUserPool
{
    public class MyCognitoUserPoolStack : Stack
    {
        internal MyCognitoUserPoolStack(Construct scope, string id, IStackProps props = null) : base(scope, id, props)
        {
            // The code that defines your stack goes here
            var userpool = new UserPool(this, "myuserpool", new UserPoolProps {
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

            userpool.AddClient("MyUserPoolClient", new UserPoolClientProps {});
        }
    }
}
