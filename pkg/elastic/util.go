/*
 * @FilePath: /pkg/elastic/util.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-13 10:27:08
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-13 14:11:54
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */

package elastic

import (
	"fmt"

	"github.com/elastic/go-ucfg"
	esv1 "github.com/kbsonlong/es-operator/api/v1"
	k8scorev1 "k8s.io/api/core/v1"
)

// the name of the ES attribute indicating the pod's current k8s node
const nodeAttrK8sNodeName = "k8s_node_name"

var nodeAttrNodeName = fmt.Sprintf("%s.%s", esv1.NodeAttr, nodeAttrK8sNodeName)
var Options = []ucfg.Option{ucfg.PathSep("."), ucfg.AppendValues}

// baseConfig returns the base ES configuration to apply for the given cluster
func BaseConfig(clusterName string, ipFamily k8scorev1.IPFamily) *ucfg.Config {
	cfg := map[string]interface{}{
		// derive node name dynamically from the pod name, injected as env var
		esv1.NodeName:    "${POD_NAME}",
		esv1.ClusterName: clusterName,

		// use the DNS name as the publish host
		esv1.NetworkPublishHost: IPLiteralFor("${POD_IP}", ipFamily),
		esv1.HTTPPublishHost:    "${POD_NAME}.${HEADLESS_SERVICE_NAME}.${NAMESPACE}.svc",
		esv1.NetworkHost:        "0",

		// allow ES to be aware of k8s node the pod is running on when allocating shards
		esv1.ShardAwarenessAttributes: nodeAttrK8sNodeName,
		nodeAttrNodeName:              "${NODE_NAME}",

		// esv1.PathData: volume.ElasticsearchDataMountPath,
		// esv1.PathLogs: volume.ElasticsearchLogsMountPath,
	}

	// seed hosts setting name changed starting ES 7.X
	// fileProvider := "file"
	// if ver.Major < 7 {
	// 	cfg[esv1.DiscoveryZenHostsProvider] = fileProvider
	// } else {
	// 	cfg[esv1.DiscoverySeedProviders] = fileProvider
	// 	// to avoid misleading error messages about the inability to connect to localhost for discovery despite us using
	// 	// file based discovery
	// 	cfg[esv1.DiscoverySeedHosts] = []string{}
	// }
	basecfg, _ := ucfg.NewFrom(cfg)

	return basecfg
}

func IPLiteralFor(ipOrPlaceholder string, ipFamily k8scorev1.IPFamily) string {
	if ipFamily == k8scorev1.IPv6Protocol {
		// IPv6: return a bracketed version of the IP
		return fmt.Sprintf("[%s]", ipOrPlaceholder)
	}
	// IPv4: leave the placeholder as is
	return ipOrPlaceholder
}
