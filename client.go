// Copyright 2019 The go-language-server Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import (
	"context"

	"github.com/francoispqt/gojay"
	"github.com/go-language-server/jsonrpc2"
	"go.uber.org/zap"
)

// ClientInterface represents a Language Server Protocol client.
type ClientInterface interface {
	Run(ctx context.Context) (err error)
	LogMessage(ctx context.Context, params *LogMessageParams) (err error)
	PublishDiagnostics(ctx context.Context, params *PublishDiagnosticsParams) (err error)
	ShowMessage(ctx context.Context, params *ShowMessageParams) (err error)
	ShowMessageRequest(ctx context.Context, params *ShowMessageRequestParams) (result *MessageActionItem, err error)
	Telemetry(ctx context.Context, params interface{}) (err error)
	RegisterCapability(ctx context.Context, params *RegistrationParams) (err error)
	UnregisterCapability(ctx context.Context, params *UnregistrationParams) (err error)
	WorkspaceApplyEdit(ctx context.Context, params *ApplyWorkspaceEditParams) (result bool, err error)
	WorkspaceConfiguration(ctx context.Context, params *ConfigurationParams) (result []interface{}, err error)
	WorkspaceFolders(ctx context.Context) (result []WorkspaceFolder, err error)
}

const (
	MethodWindowShowMessage              = "window/showMessage"
	MethodWindowShowMessageRequest       = "window/showMessageRequest"
	MethodWindowLogMessage               = "window/logMessage"
	MethodTelemetryEvent                 = "telemetry/event"
	MethodClientRegisterCapability       = "client/registerCapability"
	MethodClientUnregisterCapability     = "client/unregisterCapability"
	MethodTextDocumentPublishDiagnostics = "textDocument/publishDiagnostics"
	MethodWorkspaceApplyEdit             = "workspace/applyEdit"
	MethodWorkspaceConfiguration         = "workspace/configuration"
	MethodWorkspaceWorkspaceFolders      = "workspace/workspaceFolders"
)

// Client implements a Language Server Protocol client.
type Client struct {
	*jsonrpc2.Conn
	logger *zap.Logger
}

var _ ClientInterface = (*Client)(nil)

// Run runs the Language Server Protocol client.
func (c *Client) Run(ctx context.Context) (err error) {
	err = c.Conn.Run(ctx)
	return
}

// LogMessage sents the notification from the server to the client to ask the client to log a particular message.
func (c *Client) LogMessage(ctx context.Context, params *LogMessageParams) (err error) {
	err = c.Conn.Notify(ctx, MethodWindowLogMessage, params)
	return
}

// PublishDiagnostics sends the notification from the server to the client to signal results of validation runs.
//
// Diagnostics are “owned” by the server so it is the server’s responsibility to clear them if necessary. The following rule is used for VS Code servers that generate diagnostics:
//
// - if a language is single file only (for example HTML) then diagnostics are cleared by the server when the file is closed.
// - if a language has a project system (for example C#) diagnostics are not cleared when a file closes. When a project is opened all diagnostics for all files are recomputed (or read from a cache).
//
// When a file changes it is the server’s responsibility to re-compute diagnostics and push them to the client.
// If the computed set is empty it has to push the empty array to clear former diagnostics.
// Newly pushed diagnostics always replace previously pushed diagnostics. There is no merging that happens on the client side.
func (c *Client) PublishDiagnostics(ctx context.Context, params *PublishDiagnosticsParams) (err error) {
	err = c.Conn.Notify(ctx, MethodTextDocumentPublishDiagnostics, params)
	return
}

// ShowMessage sends the notification from a server to a client to ask the
// client to display a particular message in the user interface.
func (c *Client) ShowMessage(ctx context.Context, params *ShowMessageParams) (err error) {
	err = c.Conn.Notify(ctx, MethodWindowShowMessage, params)
	return
}

// ShowMessageRequest sends the request from a server to a client to ask the client to display a particular message in the user interface.
//
// In addition to the show message notification the request allows to pass actions and to wait for an answer from the client.
func (c *Client) ShowMessageRequest(ctx context.Context, params *ShowMessageRequestParams) (result *MessageActionItem, err error) {
	result = new(MessageActionItem)
	err = c.Conn.Call(ctx, MethodWindowShowMessageRequest, params, result)

	return result, err
}

// Telemetry sents the notification from the server to the client to ask the client to log a telemetry event.
func (c *Client) Telemetry(ctx context.Context, params interface{}) (err error) {
	err = c.Conn.Notify(ctx, MethodTelemetryEvent, params)
	return
}

// RegisterCapability sents the request from the server to the client to register for a new capability on the client side.
//
// Not all clients need to support dynamic capability registration.
//
// A client opts in via the dynamicRegistration property on the specific client capabilities.
// A client can even provide dynamic registration for capability A but not for capability B (see TextDocumentClientCapabilities as an example).
func (c *Client) RegisterCapability(ctx context.Context, params *RegistrationParams) (err error) {
	err = c.Conn.Notify(ctx, MethodClientRegisterCapability, params)
	return
}

// UnregisterCapability sents the request from the server to the client to unregister a previously registered capability.
func (c *Client) UnregisterCapability(ctx context.Context, params *UnregistrationParams) (err error) {
	err = c.Conn.Notify(ctx, MethodClientUnregisterCapability, params)
	return
}

// WorkspaceApplyEdit sends the request from the server to the client to modify resource on the client side.
func (c *Client) WorkspaceApplyEdit(ctx context.Context, params *ApplyWorkspaceEditParams) (result bool, err error) {
	err = c.Conn.Call(ctx, MethodWorkspaceApplyEdit, params, &result)

	return result, err
}

// WorkspaceConfiguration sends the request from the server to the client to fetch configuration settings from the client.
//
// The request can fetch several configuration settings in one roundtrip.
// The order of the returned configuration settings correspond to the order of the
// passed ConfigurationItems (e.g. the first item in the response is the result for the first configuration item in the params).
func (c *Client) WorkspaceConfiguration(ctx context.Context, params *ConfigurationParams) ([]interface{}, error) {
	var result []interface{}
	err := c.Conn.Call(ctx, MethodWorkspaceConfiguration, params, &result)

	return result, err
}

// WorkspaceFolders sents the request from the server to the client to fetch the current open list of workspace folders.
//
// Returns null in the response if only a single file is open in the tool. Returns an empty array if a workspace is open but no folders are configured.
//
// Since version 3.6.0.
func (c *Client) WorkspaceFolders(ctx context.Context) (result []WorkspaceFolder, err error) {
	err = c.Conn.Call(ctx, MethodWorkspaceWorkspaceFolders, nil, &result)

	return result, err
}

// ClientHandler returns the client handler.
func ClientHandler(client ClientInterface, logger *zap.Logger) jsonrpc2.Handler {
	return func(ctx context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) {
		logger.Debug("ClientHandler", zap.String("r.Method", r.Method))

		switch r.Method {
		case MethodCancelRequest:
			var params CancelParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}
			conn.Cancel(params.ID)

		case MethodClientRegisterCapability:
			var params RegistrationParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			if err := client.RegisterCapability(ctx, &params); err != nil {
				logger.Error(MethodClientRegisterCapability, zap.Error(err))
			}

		case MethodClientUnregisterCapability:
			var params UnregistrationParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			if err := client.UnregisterCapability(ctx, &params); err != nil {
				logger.Error(MethodClientUnregisterCapability, zap.Error(err))
			}

		case MethodTelemetryEvent:
			var params interface{}
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			if err := client.Telemetry(ctx, &params); err != nil {
				logger.Error(MethodTelemetryEvent, zap.Error(err))
			}

		case MethodTextDocumentPublishDiagnostics:
			var params PublishDiagnosticsParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			if err := client.PublishDiagnostics(ctx, &params); err != nil {
				logger.Error(MethodTextDocumentPublishDiagnostics, zap.Error(err))
			}

		case MethodWindowLogMessage:
			var params LogMessageParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			if err := client.LogMessage(ctx, &params); err != nil {
				logger.Error(MethodWindowLogMessage, zap.Error(err))
			}

		case MethodWindowShowMessage:
			var params ShowMessageParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			if err := client.ShowMessage(ctx, &params); err != nil {
				logger.Error(MethodWindowShowMessage, zap.Error(err))
			}

		case MethodWindowShowMessageRequest:
			var params ShowMessageRequestParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			resp, err := client.ShowMessageRequest(ctx, &params)
			if err := conn.Reply(ctx, r, resp, err); err != nil {
				logger.Error(MethodWindowShowMessageRequest, zap.Error(err))
			}

		case MethodWorkspaceApplyEdit:
			var params ApplyWorkspaceEditParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			resp, err := client.WorkspaceApplyEdit(ctx, &params)
			if err := conn.Reply(ctx, r, resp, err); err != nil {
				logger.Error(MethodWorkspaceApplyEdit, zap.Error(err))
			}

		case MethodWorkspaceConfiguration:
			var params ConfigurationParams
			if err := gojay.Unsafe.Unmarshal(*r.Params, &params); err != nil {
				ReplyError(ctx, err, conn, r, logger)
				return
			}

			resp, err := client.WorkspaceConfiguration(ctx, &params)
			if err := conn.Reply(ctx, r, resp, err); err != nil {
				logger.Error(MethodWorkspaceConfiguration, zap.Error(err))
			}

		case MethodWorkspaceWorkspaceFolders:
			if r.Params != nil {
				conn.Reply(ctx, r, nil, jsonrpc2.Errorf(jsonrpc2.CodeInvalidParams, "Expected no params"))
				return
			}

			resp, err := client.WorkspaceFolders(ctx)
			if err := conn.Reply(ctx, r, resp, err); err != nil {
				logger.Error(MethodWorkspaceWorkspaceFolders, zap.Error(err))
			}

		default:
			if r.IsNotify() {
				conn.Reply(ctx, r, nil, jsonrpc2.Errorf(jsonrpc2.CodeMethodNotFound, "method %q not found", r.Method))
			}
		}
	}
}
