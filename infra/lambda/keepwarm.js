let https = require('https')

exports.handler = (event, context, callback) => {
	let apiDomain = 'https://' + process.env.API_DOMAIN + '/keepwarm'
	https.get(apiDomain, (res) => {
		callback(null, res.statusCode)
	}).on('error', (e) => {
		callback(Error(e))
	})
}