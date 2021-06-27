package sms

func errorCodeToMessage(code int) string {
	switch code {
	case 1:
		return "You have exceeded the submission capacity allowed on this account. Please wait and retry."
	case 2:
		return "Your request is incomplete and missing some mandatory parameters."
	case 3:
		return "The value of one or more parameters is invalid."
	case 4:
		return "The api_key / api_secret you supplied is either invalid or disabled."
	case 5:
		return "There was an error processing your request in the Platform."
	case 6:
		return "The Platform was unable to process your request. For example, due to an unrecognised prefix for the phone number."
	case 7:
		return "The number you are trying to submit to is blacklisted and may not receive messages."
	case 8:
		return "The api_key you supplied is for an account that has been barred from submitting messages."
	case 9:
		return "Your pre-paid account does not have sufficient credit to process this message."
	case 11:
		return "This account is not provisioned for REST submission, you should use SMPP instead."
	case 12:
		return "The length of udh and body was greater than 140 octets for a binary type SMS request."
	case 13:
		return "Message was not submitted because there was a communication failure."
	case 14:
		return "Message was not submitted due to a verification failure in the submitted signature."
	case 15:
		return "Due to local regulations, the SenderID you set in from in the request was not accepted. Please check the Global messaging section."
	case 16:
		return "The value of ttl in your request was invalid."
	case 19:
		return "Your request makes use of a facility that is not enabled on your account."
	case 20:
		return "The value of message-class in your request was out of range. See https://en.wikipedia.org/wiki/Data_Coding_Scheme."
	case 23:
		return "You did not include https in the URL you set in callback."
	case 29:
		return "The phone number you set in to is not in your pre-approved destination list. To send messages to this phone number, add it using Dashboard."
	case 34:
		return "The phone number you supplied in the to parameter of your request was either missing or invalid."
	}

	return ""
}
