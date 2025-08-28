/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 *
 * Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *      http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package logger

import (
	"context"
	"fmt"
	"log"
)

var logger Logger

func init() {
	logger = &ConsoleLogger{}
}

// SetLogger - set custom logger
func SetLogger(customLogger Logger) {
	logger = customLogger
}

// Logger logging interface for gonvme
type Logger interface {
	Info(ctx context.Context, format string, args ...interface{})
	Debug(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, format string, args ...interface{})
}

// ConsoleLogger - placeholder for default logger
type ConsoleLogger struct{}

// Info - log info using default logger
func (dl *ConsoleLogger) Info(_ context.Context, format string, args ...interface{}) {
	log.Print("INFO: " + fmt.Sprintf(format, args...))
}

// Debug - log debug using default logger
func (dl *ConsoleLogger) Debug(_ context.Context, format string, args ...interface{}) {
	log.Print("DEBUG: " + fmt.Sprintf(format, args...))
}

// Error - log error using default logger
func (dl *ConsoleLogger) Error(_ context.Context, format string, args ...interface{}) {
	log.Print("ERROR: " + fmt.Sprintf(format, args...))
}

// Info - log info using custom logger
func Info(ctx context.Context, format string, args ...interface{}) {
	logger.Info(ctx, format, args...)
}

// Debug - log debug using custom logger
func Debug(ctx context.Context, format string, args ...interface{}) {
	logger.Debug(ctx, format, args...)
}

// Error - log error using custom logger
func Error(ctx context.Context, format string, args ...interface{}) {
	logger.Error(ctx, format, args...)
}
