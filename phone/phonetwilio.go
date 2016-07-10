package phone

//TwilioConfig represents all necessary information needed to send a message via
// twilio. For more information, see http://twilio.com
type TwilioConfig struct {
	SID       string
	AuthToken string
	SenderNum string
}
