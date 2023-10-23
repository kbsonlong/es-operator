/*
 * @FilePath: /api/v1/elasticsearch_config.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-13 10:28:40
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-13 10:30:48
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */
package v1

import "github.com/elastic/go-ucfg"

const (
	ClusterName = "cluster.name"

	DiscoveryZenMinimumMasterNodes = "discovery.zen.minimum_master_nodes"
	ClusterInitialMasterNodes      = "cluster.initial_master_nodes"

	DiscoveryZenHostsProvider = "discovery.zen.hosts_provider" // ES < 7.X
	DiscoverySeedProviders    = "discovery.seed_providers"     // ES >= 7.X
	DiscoverySeedHosts        = "discovery.seed_hosts"         // ES >= 7.X

	NetworkHost        = "network.host"
	NetworkPublishHost = "network.publish_host"
	HTTPPublishHost    = "http.publish_host"

	NodeName = "node.name"

	PathData = "path.data"
	PathLogs = "path.logs"

	ShardAwarenessAttributes = "cluster.routing.allocation.awareness.attributes"
	NodeAttr                 = "node.attr"

	XPackSecurityAuthcRealmsFileFile1Order     = "xpack.security.authc.realms.file.file1.order"     // 7.x realm syntax
	XPackSecurityAuthcRealmsFile1Order         = "xpack.security.authc.realms.file1.order"          // 6.x realm syntax
	XPackSecurityAuthcRealmsFile1Type          = "xpack.security.authc.realms.file1.type"           // 6.x realm syntax
	XPackSecurityAuthcRealmsNativeNative1Order = "xpack.security.authc.realms.native.native1.order" // 7.x realm syntax
	XPackSecurityAuthcRealmsNative1Order       = "xpack.security.authc.realms.native1.order"        // 6.x realm syntax
	XPackSecurityAuthcRealmsNative1Type        = "xpack.security.authc.realms.native1.type"         // 6.x realm syntax

	XPackSecurityAuthcReservedRealmEnabled          = "xpack.security.authc.reserved_realm.enabled"
	XPackSecurityEnabled                            = "xpack.security.enabled"
	XPackSecurityHttpSslCertificate                 = "xpack.security.http.ssl.certificate"             //nolint:revive
	XPackSecurityHttpSslCertificateAuthorities      = "xpack.security.http.ssl.certificate_authorities" //nolint:revive
	XPackSecurityHttpSslClientAuthentication        = "xpack.security.http.ssl.client_authentication"   //nolint:revive
	XPackSecurityHttpSslEnabled                     = "xpack.security.http.ssl.enabled"                 //nolint:revive
	XPackSecurityHttpSslKey                         = "xpack.security.http.ssl.key"                     //nolint:revive
	XPackSecurityTransportSslCertificate            = "xpack.security.transport.ssl.certificate"
	XPackSecurityTransportSslCertificateAuthorities = "xpack.security.transport.ssl.certificate_authorities"
	XPackSecurityTransportSslEnabled                = "xpack.security.transport.ssl.enabled"
	XPackSecurityTransportSslKey                    = "xpack.security.transport.ssl.key"
	XPackSecurityTransportSslVerificationMode       = "xpack.security.transport.ssl.verification_mode"

	XPackLicenseUploadTypes = "xpack.license.upload.types" // supported >= 7.6.0 used as of 7.8.1
)

type Config struct {
	// Data holds the configuration keys and values.
	// This field exists to work around https://github.com/kubernetes-sigs/kubebuilder/issues/528
	Data map[string]interface{} `json:"-"`
}

type CanonicalConfig struct {
	*ucfg.Config
}
