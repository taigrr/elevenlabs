package elevenlabs

func convaiEndpointCases() []endpointCase {
	cases := convaiAgentEndpointCases()
	cases = append(cases, convaiConversationEndpointCases()...)
	return cases
}
