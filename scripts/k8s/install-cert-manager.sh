#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}cert-manager Installation Script${NC}"
echo "=================================="

# Function to compare versions
# Returns: 0 if versions are equal, 1 if v1 > v2, 2 if v1 < v2
version_compare() {
    local v1=$1
    local v2=$2
    
    # Remove 'v' prefix if present
    v1=${v1#v}
    v2=${v2#v}
    
    if [[ "$v1" == "$v2" ]]; then
        return 0
    fi
    
    # Split versions into arrays
    IFS='.' read -ra V1 <<< "$v1"
    IFS='.' read -ra V2 <<< "$v2"
    
    # Compare each part
    local max_length=${#V1[@]}
    if [[ ${#V2[@]} -gt $max_length ]]; then
        max_length=${#V2[@]}
    fi
    
    for ((i=0; i<max_length; i++)); do
        local part1=${V1[i]:-0}
        local part2=${V2[i]:-0}
        
        if [[ $part1 -gt $part2 ]]; then
            return 1
        elif [[ $part1 -lt $part2 ]]; then
            return 2
        fi
    done
    
    return 0
}

# Function to get latest cert-manager version from GitHub
get_latest_version() {
    # Send informational messages to stderr to avoid contaminating the return value
    echo -e "${BLUE}Fetching latest cert-manager version...${NC}" >&2
    
    # Try multiple methods to get the latest release
    local latest_version=""
    local method_used=""
    
    # Method 1: Use GitHub API with python3 parsing (most reliable)
    if command -v python3 &> /dev/null; then
        latest_version=$(curl -s https://api.github.com/repos/cert-manager/cert-manager/releases/latest 2>/dev/null | \
            python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    print(data['tag_name'])
except:
    pass
" 2>/dev/null)
        
        if [[ -n "$latest_version" && "$latest_version" != "null" ]]; then
            method_used="GitHub API (python3)"
        fi
    fi
    
    # Method 2: Fallback to grep/sed if python3 method failed
    if [[ -z "$latest_version" || "$latest_version" == "null" ]]; then
        latest_version=$(curl -s https://api.github.com/repos/cert-manager/cert-manager/releases/latest 2>/dev/null | \
            grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' 2>/dev/null)
        
        if [[ -n "$latest_version" && "$latest_version" != "null" ]]; then
            method_used="GitHub API (grep/sed)"
        fi
    fi
    
    # Method 3: Try alternative GitHub API endpoint
    if [[ -z "$latest_version" || "$latest_version" == "null" ]]; then
        if command -v python3 &> /dev/null; then
            latest_version=$(curl -s https://api.github.com/repos/cert-manager/cert-manager/tags 2>/dev/null | \
                python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    # Find the first stable release (not pre-release)
    for tag in data[:10]:  # Check first 10 tags
        version = tag['name']
        if version and not ('alpha' in version or 'beta' in version or 'rc' in version):
            print(version)
            break
except:
    pass
" 2>/dev/null)
            
            if [[ -n "$latest_version" && "$latest_version" != "null" ]]; then
                method_used="GitHub Tags API (python3)"
            fi
        fi
    fi
    
    # Validate the version format
    if [[ -n "$latest_version" && "$latest_version" != "null" ]]; then
        # Check if it looks like a valid version (starts with v and contains dots)
        if [[ "$latest_version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
            echo -e "${GREEN}Latest cert-manager version: ${latest_version}${NC}" >&2
            echo -e "${BLUE}Detection method: ${method_used}${NC}" >&2
            
            # Verify with recent releases (optional validation)
            if command -v python3 &> /dev/null; then
                echo -e "${BLUE}Verifying with recent releases...${NC}" >&2
                recent_releases=$(curl -s https://api.github.com/repos/cert-manager/cert-manager/releases?per_page=3 2>/dev/null | \
                    python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    releases = []
    for r in data[:3]:
        status = 'pre-release' if r.get('prerelease', False) else 'stable'
        releases.append(f'  {r[\"tag_name\"]} ({status})')
    if releases:
        print('Recent releases:')
        print('\\n'.join(releases))
except:
    pass
" 2>/dev/null)
                
                if [[ -n "$recent_releases" ]]; then
                    echo -e "${YELLOW}${recent_releases}${NC}" >&2
                fi
            fi
            
            # Return only the version number to stdout
            echo "$latest_version"
            return
        else
            echo -e "${YELLOW}Warning: Invalid version format detected: ${latest_version}${NC}" >&2
        fi
    fi
    
    # All methods failed, use fallback
    echo -e "${YELLOW}Warning: Could not fetch latest version from GitHub API${NC}" >&2
    echo -e "${YELLOW}This could be due to network issues or rate limiting${NC}" >&2
    
    # Use a known stable version as fallback
    latest_version="v1.16.1"
    echo -e "${YELLOW}Falling back to known stable version ${latest_version}${NC}" >&2
    echo -e "${YELLOW}You may want to check https://github.com/cert-manager/cert-manager/releases manually${NC}" >&2
    # Return only the version number to stdout
    echo "$latest_version"
}

# Function to get installed cert-manager version
get_installed_version() {
    if kubectl get crd certificates.cert-manager.io &> /dev/null; then
        local installed_version
        installed_version=$(kubectl get deploy -n cert-manager cert-manager -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null | cut -d: -f2)
        
        if [[ -n "$installed_version" ]]; then
            echo "$installed_version"
        else
            echo ""
        fi
    else
        echo ""
    fi
}

# Function to install cert-manager
install_cert_manager() {
    local version=$1
    echo -e "${BLUE}Installing cert-manager ${version}...${NC}"
    
    kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/${version}/cert-manager.yaml"
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}Failed to apply cert-manager manifest${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}Waiting for cert-manager to be ready...${NC}"
    kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager -n cert-manager
    kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager-webhook -n cert-manager
    kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager-cainjector -n cert-manager
    
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}cert-manager ${version} installed successfully!${NC}"
    else
        echo -e "${RED}cert-manager installation failed or timed out${NC}"
        exit 1
    fi
}

# Function to verify installation
verify_installation() {
    echo -e "${BLUE}Verifying cert-manager installation...${NC}"
    kubectl get pods --namespace cert-manager
    
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}cert-manager is running successfully!${NC}"
    else
        echo -e "${RED}cert-manager verification failed${NC}"
        exit 1
    fi
}

# Main script logic
main() {
    # Get latest version
    latest_version=$(get_latest_version)
    
    # Check if cert-manager is already installed
    installed_version=$(get_installed_version)
    
    if [[ -z "$installed_version" ]]; then
        # cert-manager is not installed
        echo -e "${YELLOW}cert-manager is not installed${NC}"
        echo -e "${BLUE}Installing latest version: ${latest_version}${NC}"
        install_cert_manager "$latest_version"
        verify_installation
    else
        # cert-manager is already installed
        echo -e "${GREEN}cert-manager is already installed${NC}"
        echo -e "${BLUE}Installed version: ${installed_version}${NC}"
        echo -e "${BLUE}Latest version: ${latest_version}${NC}"
        
        # Compare versions
        version_compare "$installed_version" "$latest_version"
        local comparison=$?
        
        case $comparison in
            0)
                echo -e "${GREEN}You have the latest version installed!${NC}"
                echo -e "${BLUE}No update needed.${NC}"
                ;;
            1)
                echo -e "${YELLOW}You have a newer version than the latest release!${NC}"
                echo -e "${YELLOW}This might be a pre-release or development version.${NC}"
                echo -e "${BLUE}No action needed.${NC}"
                ;;
            2)
                echo -e "${YELLOW}A newer version is available!${NC}"
                echo ""
                read -p "Do you want to update from ${installed_version} to ${latest_version}? (y/N) " -n 1 -r
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    echo -e "${BLUE}Updating cert-manager...${NC}"
                    install_cert_manager "$latest_version"
                    verify_installation
                else
                    echo -e "${BLUE}Update cancelled. Keeping current version.${NC}"
                fi
                ;;
        esac
    fi
    
    echo ""
    echo -e "${GREEN}cert-manager setup complete!${NC}"
}

# Check prerequisites
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed or not in PATH${NC}"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is not installed or not in PATH${NC}"
    exit 1
fi

# Run main function
main 