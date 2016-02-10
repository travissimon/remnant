# Remnant

Microservices observability and logging, heavily inspired by Google's Dapper paper and Twitter's Zipkin implementation.


Client links:

[{
	name: "log message",
	spanId: "string",
	severity: "string ('debug'|'info'|'warning'|'error')",
	message: "string",
	stackTrace: "string (for error/exceptions)"
}, {
	name: "client start"
	url: "/client/start",
	actions: ["POST"],
	body: {
		traceId: "string",
		spanId: "string",
		parentSpanId: "string",
		ipAddress: "string",
		requestUrl: "string",
		requestParameters: [{key: string, value: string}]
	},
}, {
	name: "client end"
	url: "/client/end",
	actions: ["POST"], 
	body: {
		traceId: "string",
		spanId: "string",
		parentSpanId: "string",
	},
}, {
	name: "remote call start"
	url: "/remote-call/start",
	actions: ["POST"],
	body: {
		traceId: "string",
		spanId: "string",
		parentSpanId: "string",
	},
}, {
	name: "remote call end"
	url: "/remote-call/end",
	actions: ["POST"], 
	body: {
		traceId: "string",
		spanId: "string",
		parentSpanId: "string",
		responseCode: "http status code"
	},
}]

Server Urls:
[{
	name: "most recent",
	actions: ["GET"],
	template: "/recent/{count}",
}, {
	name: "most recent by ip",
	actions: ["GET"],
	template: "/recent-by-ip/{ip-address}/{count}",
}]

Note: most recent should return:
{
	traces: [{
		url: "/trace/1",
		ip: "127.0.0.1",
		start: "2016-02-06T20:37:47Z",
		end: "2016-02-06T20:37:58Z"
	}, {
		url: "/trace/2",
		ip: "127.0.0.1",
		start: "2016-02-06T20:35:02Z",
		end: "2016-02-06T20:35:28Z"
	}],
	links: [{
		name: "next",
		url: "/recent/10/offset/20",
	}(optional: , {
		name: "previous",
		url: "/recent/10/offset/0",
	})]
}]

Trace returns:
{
	traceId: "string",
	start: "2016-02-06T20:35:02Z",
	end: "2016-02-06T20:35:28Z",
	spans: [{
	}]

}


Queries - use original call for info
- 10 most recent
- 10 most recent filtered by ip/name
- Index listingget traceid, ip/name, function called


Access by TraceId
- Retreive trace data and all span data
