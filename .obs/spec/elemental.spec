#
# spec file for package elemental
#
# Copyright (c) 2022 SUSE LLC
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via https://bugs.opensuse.org/
#

%define systemdir /system
%define oemdir %{systemdir}/oem

Name:           elemental
Version:        0
Release:        0
Summary:        A Rancher and Kubernetes optimized immutable Linux distribution
License:        Apache-2.0
Group:          System/Management
URL:            https://github.com/rancher-sandbox/%{name}
Source:         %{name}-%{version}.tar
Source1:        LICENSE
Source2:        README.md

Requires:       elemental-toolkit
Requires:       elemental-register
Requires:       elemental-system-agent
Requires:       elemental-support
Requires:       NetworkManager
Requires:       systemd-presets-branding-Elemental
Requires:       elemental-updater = %{version}-%{release}
%{?systemd_requires}

BuildArch:      noarch
BuildRoot:      %{_tmppath}/%{name}-%{version}-build

%description
Elemental is a new set of tools to manage operating systems as container images within Rancher Manager

%package -n elemental-updater
Summary:        Rancher elemental node updater
Group:          System/Management

%description -n elemental-updater
Rancher elemental node updater. To be installed on the node.

%prep
%setup -q -n %{name}-%{version}
cp %{S:1} .
cp %{S:2} .

%build


%install

cp -a framework/files/* %{buildroot}

rm -rf %{buildroot}/var/log/journal

# remove luet config in Elemental Teal
rm -rf %{buildroot}/etc/luet 

# belongs to elemental-system-agent package
rm %{buildroot}%{_unitdir}/elemental-system-agent.service

# remove placeholders
rm -rf %{buildroot}/usr/libexec/.placeholder

%pre
%service_add_pre elemental-populate-node-labels.service

%post
%service_add_post elemental-populate-node-labels.service

%preun
%service_del_preun elemental-populate-node-labels.service

%postun
%service_del_postun elemental-populate-node-labels.service

%files
%defattr(-,root,root,-)
%doc README.md
%license LICENSE
%dir %{_sysconfdir}/cos
%config %{_sysconfdir}/cos/bootargs.cfg
%dir %{_sysconfdir}/dracut.conf.d
%config %{_sysconfdir}/dracut.conf.d/51-certificates-initrd.conf
%config %{_sysconfdir}/dracut.conf.d/99-teal-systemd.conf
%dir %{_sysconfdir}/NetworkManager
%dir %{_sysconfdir}/NetworkManager/conf.d
%config %{_sysconfdir}/NetworkManager/conf.d/rke2-canal.conf
%dir %{_unitdir}
%{_unitdir}/elemental-populate-node-labels.service
%{_sbindir}/elemental-populate-node-labels
%dir /usr/libexec
%dir %{systemdir}
%dir %{oemdir}
%{oemdir}/*

%files -n elemental-updater
%defattr(-,root,root,-)
%license LICENSE
%{_sbindir}/self-upgrade
%{_sbindir}/suc-upgrade

%changelog
