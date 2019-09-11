using System;
using Grpc.Core;
using Getcd;

namespace aevm
{
    public class gRPCClient
    {
        protected string _host = "";
        protected Channel _channel = null;
        protected GetcdService.GetcdServiceClient _client = null;
        
        public gRPCClient()
        {
            
        }

        public GetcdService.GetcdServiceClient getcd
        {
            get
            {
                return _client;
            }
        }

        public bool connect(string host)
        {
            _host = host;
            _channel = new Channel(host, ChannelCredentials.Insecure);
            _client = new GetcdService.GetcdServiceClient(_channel);

            return true;
        }
    }
}
