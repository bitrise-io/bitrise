package corelog

//func Test_GivenLEgacyLogger_WhenLogMessageInvoked_ThenLogsItCorrectly(t *testing.T) {
//	tests := []struct {
//		name                string
//		hasOutput           bool
//		parameters          testLogParameters
//		expectedLogFunction string
//		expectedMessage     string
//	}{
//		{
//			name:      "Error log",
//			hasOutput: true,
//			parameters: testLogParameters{
//				producer: Step,
//				level:    ErrorLevel,
//				message:  "Error",
//			},
//			expectedLogFunction: "Errorf",
//			expectedMessage:     "Error",
//		},
//		{
//			name:      "Warning log",
//			hasOutput: true,
//			parameters: testLogParameters{
//				producer: Step,
//				level:    WarnLevel,
//				message:  "Warning",
//			},
//			expectedLogFunction: "Warnf",
//			expectedMessage:     "Warning",
//		},
//		{
//			name:      "Info log",
//			hasOutput: true,
//			parameters: testLogParameters{
//				producer: BitriseCLI,
//				level:    InfoLevel,
//				message:  "Info",
//			},
//			expectedLogFunction: "Infof",
//			expectedMessage:     "Info",
//		},
//		{
//			name:      "Done log",
//			hasOutput: true,
//			parameters: testLogParameters{
//				producer: BitriseCLI,
//				level:    DoneLevel,
//				message:  "Done",
//			},
//			expectedLogFunction: "Donef",
//			expectedMessage:     "Done",
//		},
//		{
//			name:      "Normal log",
//			hasOutput: true,
//			parameters: testLogParameters{
//				producer: Step,
//				level:    NormalLevel,
//				message:  "Normal",
//			},
//			expectedLogFunction: "Printf",
//			expectedMessage:     "Normal",
//		},
//		{
//			name:      "Debug log",
//			hasOutput: true,
//			parameters: testLogParameters{
//				producer: Step,
//				level:    DebugLevel,
//				message:  "Debug",
//			},
//			expectedLogFunction: "Debugf",
//			expectedMessage:     "Debug",
//		},
//		{
//			name:      "Debug log is not logged when disabled",
//			hasOutput: false,
//			parameters: testLogParameters{
//				producer: Step,
//				level:    DebugLevel,
//				message:  "Debug",
//			},
//			expectedLogFunction: "Debugf",
//			expectedMessage:     "Debug",
//		},
//		{
//			name:      "Empty message is logged",
//			hasOutput: true,
//			parameters: testLogParameters{
//				producer: BitriseCLI,
//				level:    InfoLevel,
//				message:  "\n",
//			},
//			expectedLogFunction: "Infof",
//			expectedMessage:     "\n",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			mockLogger := &mocks.Logger{}
//			mockLogger.On(tt.expectedLogFunction, mock.Anything).Return()
//			mockLogger.On("EnableDebugLog", mock.Anything).Return()
//
//			logger := newConsoleLogger(mockLogger)
//			logger.LogMessage(tt.parameters.producer, tt.parameters.level, tt.parameters.message)
//
//			if tt.hasOutput {
//				mockLogger.AssertCalled(t, tt.expectedLogFunction, tt.expectedMessage)
//			} else {
//				mockLogger.AssertNotCalled(t, tt.expectedLogFunction, tt.expectedMessage)
//			}
//		})
//	}
//}
