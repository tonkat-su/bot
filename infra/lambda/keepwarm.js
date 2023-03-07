const https = require('https')

exports.handler = (event, context, callback) => {
	let apiDomain = process.env.API_DOMAIN
	const u = new URL('/keepwarm', apiDomain)
	https.get(u, (res) => {
		callback(null, res.statusCode)
	}).on('error', (e) => {
		callback(Error(e))
	})
}