import subprocess
import xml.etree.ElementTree as ET


def run_scan(host: str, options: str) -> dict:
    """
    Runs an Nmap scan with the given host and options.

    Returns the parsed XML output or an error message.

    Args:
        host: The target host (IP address or hostname).
        options: Nmap command-line options.

    Returns:
        A dictionary representing the parsed Nmap XML output,
        or an error message.
    """
    try:
        # Ensure XML output option is included for parsing
        if "-oX -" not in options:
            options += " -oX -"  # Add XML output to stdout

        command = f"nmap {options} {host}"
        process = subprocess.Popen(
            command.split(),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        stdout, stderr = process.communicate()

        if process.returncode != 0:
            return {"error": f"Nmap execution failed: {stderr.decode()}"}

        if not stdout:
            msg = "Nmap produced no output. Check command and installation."
            return {"error": msg}

        # Attempt to parse XML output
        try:
            root = ET.fromstring(stdout)
            return _parse_nmap_xml(root)
        except ET.ParseError:
            msg = "Failed to parse Nmap XML output. Ensure Nmap outputs XML."
            return {"error": msg}

    except FileNotFoundError:
        msg = "Nmap command not found. Ensure Nmap is installed and in PATH."
        return {"error": msg}
    except Exception as e:
        return {"error": f"An unexpected error occurred: {str(e)}"}


def _parse_nmap_xml(root: ET.Element) -> dict:
    """
    Parses the Nmap XML ET.Element object into a Python dictionary.

    This is a basic parser and might need to be more comprehensive
    depending on the required Nmap output details.
    """
    parsed_output = {}

    # Get scan information
    parsed_output['scan_args'] = root.get('args')
    parsed_output['start_time'] = root.get('startstr')

    hosts = []
    for host_element in root.findall('host'):
        host_data = {
            'status': host_element.find('status').get('state'),
            'addresses': {},
            'hostnames': [],
            'ports': []
        }

        # Get addresses
        addr_elements = host_element.findall('address')
        for addr_el in addr_elements:
            host_data['addresses'][addr_el.get('addrtype')] = addr_el.get('addr')

        # Get hostnames
        hostnames_element = host_element.find('hostnames')
        if hostnames_element is not None:
            for hostname_element in hostnames_element.findall('hostname'):
                host_data['hostnames'].append({
                    'name': hostname_element.get('name'),
                    'type': hostname_element.get('type')
                })

        # Get port information
        ports_element = host_element.find('ports')
        if ports_element is not None:
            for port_element in ports_element.findall('port'):
                port_data = {
                    'protocol': port_element.get('protocol'),
                    'portid': port_element.get('portid'),
                    'state': port_element.find('state').get('state'),
                }
                service_el = port_element.find('service')
                if service_el:
                    name = service_el.get('name')
                    prod = service_el.get('product')
                    ver = service_el.get('version')
                    port_data['service_name'] = name
                    port_data['product'] = prod
                    port_data['version'] = ver
                host_data['ports'].append(port_data)
        hosts.append(host_data)

    parsed_output['hosts'] = hosts
    return parsed_output


if __name__ == "__main__":
    # Example usage: Scan localhost with -sV (service version detection)
    target_host = "localhost"
    scan_options = "-sV"  # Service version detection

    print(f"Running Nmap scan on {target_host} with options: {scan_options}")
    scan_result = run_scan(target_host, scan_options)

    if "error" in scan_result:
        print(f"Error: {scan_result['error']}")
    else:
        print("Scan Results:")
        # Print a summary of the results
        for host_info in scan_result.get('hosts', []):
            print(f"  Host: {host_info['addresses'].get('ipv4', 'N/A')}")
            print(f"    Status: {host_info['status']}")
            if host_info['hostnames']:
                hostnames_str = ', '.join(
                    [h['name'] for h in host_info['hostnames']]
                )
                print(f"    Hostnames: {hostnames_str}")
            print("    Ports:")
            for port in host_info['ports']:
                service_info = ""
                if 'service_name' in port:
                    service_info = f" (Service: {port['service_name']}"
                    if port.get('product'):
                        service_info += f" {port['product']}"
                    if port.get('version'):
                        service_info += f" {port['version']}"
                    service_info += ")"
                port_id = port['portid']
                protocol = port['protocol']
                state = port['state']
                # Ensure line length for print output
                details = f"{state}{service_info}"
                print(f"      - Port {port_id}/{protocol}: {details}")

        # For more detailed output, you could pretty-print the entire dictionary:
        # import json
        # print(json.dumps(scan_result, indent=2))
