Name:           usm-cli
Version:        0.0.3
Release:        1%{?dist}
Summary:        UniFi Site Manager CLI for cloud API management
License:        MIT
URL:            https://github.com/dl-alexandre/UniFi-Site-Manager-CLI
Source0:        %{name}-%{version}.tar.gz

%description
Command-line interface for managing UniFi sites via cloud API.
Supports site management, device monitoring, and network configuration.

%prep
%setup -q

%build
go build -o usm ./cmd/usm

%install
mkdir -p %{buildroot}/%{_bindir}
cp usm %{buildroot}/%{_bindir}/

%files
%{_bindir}/usm
%doc README.md
%license LICENSE

%changelog
* Mon Mar 03 2025 Dalton Alexandre <dalexandre@milcgroup.info> - 0.0.3-1
- Initial RPM release
