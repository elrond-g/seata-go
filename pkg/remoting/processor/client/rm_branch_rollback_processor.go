/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"

	"github.com/seata/seata-go/pkg/common/log"
	"github.com/seata/seata-go/pkg/protocol/message"

	getty2 "github.com/seata/seata-go/pkg/remoting/getty"
	"github.com/seata/seata-go/pkg/rm"
)

func init() {
	rmBranchRollbackProcessor := &rmBranchRollbackProcessor{}
	getty2.GetGettyClientHandlerInstance().RegisterProcessor(message.MessageType_BranchRollback, rmBranchRollbackProcessor)
}

type rmBranchRollbackProcessor struct {
}

func (f *rmBranchRollbackProcessor) Process(ctx context.Context, rpcMessage message.RpcMessage) error {
	log.Infof("the rm client received  rmBranchRollback msg %#v from tc server.", rpcMessage)
	request := rpcMessage.Body.(message.BranchRollbackRequest)
	xid := request.Xid
	branchID := request.BranchId
	resourceID := request.ResourceId
	applicationData := request.ApplicationData
	log.Infof("Branch committing: xid %s, branchID %s, resourceID %s, applicationData %s", xid, branchID, resourceID, applicationData)

	status, err := rm.GetResourceManagerInstance().GetResourceManager(request.BranchType).BranchRollback(ctx, request.BranchType, xid, branchID, resourceID, applicationData)
	if err != nil {
		log.Infof("Branch commit error: %s", err.Error())
		return err
	}

	// reply commit response to tc server
	response := message.BranchRollbackResponse{
		AbstractBranchEndResponse: message.AbstractBranchEndResponse{
			Xid:          xid,
			BranchId:     branchID,
			BranchStatus: status,
		},
	}
	err = getty2.GetGettyRemotingClient().SendAsyncResponse(response)
	if err != nil {
		log.Error("BranchCommitResponse error: {%#v}", err.Error())
		return err
	}
	return nil
}
