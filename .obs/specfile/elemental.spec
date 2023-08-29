#
# spec file for package elemental
#
# Copyright (c) 2022 - 2023 SUSE LLC
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

Requires:       elemental-cli
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

# belongs to elemental-system-agent package
rm %{buildroot}%{_unitdir}/elemental-system-agent.service

# remove placeholders
rm -rf %{buildroot}/usr/libexec/.placeholder

%pre
%if 0%{?suse_version}
%service_add_pre elemental-populate-node-labels.service
%service_add_pre shutdown-containerd.service
%service_add_pre elemental-register.service
%service_add_pre elemental-register-install.service
%service_add_pre elemental-register.timer
%endif

%post
%if 0%{?suse_version}
%service_add_post elemental-populate-node-labels.service
%service_add_post shutdown-containerd.service
%service_add_post elemental-register.service
%service_add_post elemental-register-install.service
%service_add_post elemental-register.timer
%else
%systemd_post elemental-populate-node-labels.service
%systemd_post shutdown-containerd.service
%systemd_post elemental-register.service
%systemd_post elemental-register-install.service
%systemd_post elemental-register.timer
%endif

%preun
%if 0%{?suse_version}
%service_del_preun elemental-populate-node-labels.service
%service_del_preun shutdown-containerd.service
%service_del_preun elemental-register.service
%service_del_preun elemental-register-install.service
%service_del_preun elemental-register.timer
%else
%systemd_preun elemental-populate-node-labels.service
%systemd_preun shutdown-containerd.service
%systemd_preun elemental-register.service
%systemd_preun elemental-register-install.service
%systemd_preun elemental-register.timer
%endif

%postun
%if 0%{?suse_version}
%service_del_postun elemental-populate-node-labels.service
%service_del_postun shutdown-containerd.service
%service_del_postun elemental-register.service
%service_del_postun elemental-register-install.service
%service_del_postun elemental-register.timer
%else
%systemd_postun elemental-populate-node-labels.service
%systemd_postun shutdown-containerd.service
%systemd_postun elemental-register.service
%systemd_postun elemental-register-install.service
%systemd_postun elemental-register.timer
%endif

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
%{_unitdir}/shutdown-containerd.service
%{_unitdir}/elemental-register.service
%{_unitdir}/elemental-register-install.service
%{_unitdir}/elemental-register.timer
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
