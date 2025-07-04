#!/bin/bash

# Architecture Linting Script for GoForms
# Enforces Clean Architecture principles by analyzing dependencies

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LAYERS=("domain" "application" "infrastructure" "presentation")
FORBIDDEN_IMPORTS=(
    "domain.*->infrastructure.*"
    "domain.*->presentation.*"
    "application.*->presentation.*"
    "application.*->infrastructure.*"
)

# Layer definitions for clean architecture
DOMAIN_LAYER="github.com/goformx/goforms/internal/domain"
APPLICATION_LAYER="github.com/goformx/goforms/internal/application"
INFRASTRUCTURE_LAYER="github.com/goformx/goforms/internal/infrastructure"
PRESENTATION_LAYER="github.com/goformx/goforms/internal/presentation"

echo -e "${BLUE}🔍 GoForms Architecture Linting${NC}"
echo "=================================="

# Function to check if a path belongs to a layer
check_layer() {
    local path=$1
    local layer=$2

    case $layer in
        "domain")
            [[ $path == $DOMAIN_LAYER* ]]
            ;;
        "application")
            [[ $path == $APPLICATION_LAYER* ]]
            ;;
        "infrastructure")
            [[ $path == $INFRASTRUCTURE_LAYER* ]]
            ;;
        "presentation")
            [[ $path == $PRESENTATION_LAYER* ]]
            ;;
        *)
            return 1
            ;;
    esac
}

# Function to get layer name from path
get_layer_name() {
    local path=$1

    if [[ $path == $DOMAIN_LAYER* ]]; then
        echo "domain"
    elif [[ $path == $APPLICATION_LAYER* ]]; then
        echo "application"
    elif [[ $path == $INFRASTRUCTURE_LAYER* ]]; then
        echo "infrastructure"
    elif [[ $path == $PRESENTATION_LAYER* ]]; then
        echo "presentation"
    else
        echo "external"
    fi
}

# Function to check forbidden imports
check_forbidden_imports() {
    echo -e "\n${YELLOW}🔒 Checking for forbidden imports...${NC}"

    local violations=0

    # Check domain -> infrastructure
    if grep -r "github.com/goformx/goforms/internal/infrastructure" internal/domain/ >/dev/null 2>&1; then
        echo -e "${RED}❌ VIOLATION: Domain layer imports infrastructure${NC}"
        grep -r "github.com/goformx/goforms/internal/infrastructure" internal/domain/
        violations=$((violations + 1))
    fi

    # Check domain -> presentation
    if grep -r "github.com/goformx/goforms/internal/presentation" internal/domain/ >/dev/null 2>&1; then
        echo -e "${RED}❌ VIOLATION: Domain layer imports presentation${NC}"
        grep -r "github.com/goformx/goforms/internal/presentation" internal/domain/
        violations=$((violations + 1))
    fi

    # Check application -> presentation
    if grep -r "github.com/goformx/goforms/internal/presentation" internal/application/ >/dev/null 2>&1; then
        echo -e "${RED}❌ VIOLATION: Application layer imports presentation${NC}"
        grep -r "github.com/goformx/goforms/internal/presentation" internal/application/
        violations=$((violations + 1))
    fi

    # Check application -> infrastructure (except for interfaces)
    if grep -r "github.com/goformx/goforms/internal/infrastructure" internal/application/ | grep -v "interfaces" >/dev/null 2>&1; then
        echo -e "${RED}❌ VIOLATION: Application layer imports infrastructure (non-interface)${NC}"
        grep -r "github.com/goformx/goforms/internal/infrastructure" internal/application/ | grep -v "interfaces"
        violations=$((violations + 1))
    fi

    if [ $violations -eq 0 ]; then
        echo -e "${GREEN}✅ No forbidden imports found${NC}"
    fi

    return $violations
}

# Function to analyze go mod graph
analyze_dependencies() {
    echo -e "\n${YELLOW}📊 Analyzing dependency graph...${NC}"

    local violations=0

    # Generate dependency graph
    cd "$PROJECT_ROOT"
    go mod graph > /tmp/goforms_deps.txt

    # Analyze internal dependencies
    while IFS= read -r line; do
        if [[ $line == *"github.com/goformx/goforms"* ]]; then
            # Extract source and target from dependency line
            source=$(echo "$line" | cut -d' ' -f1)
            target=$(echo "$line" | cut -d' ' -f2)

            # Skip if not internal
            if [[ $source != *"github.com/goformx/goforms"* ]] || [[ $target != *"github.com/goformx/goforms"* ]]; then
                continue
            fi

            source_layer=$(get_layer_name "$source")
            target_layer=$(get_layer_name "$target")

            # Check for violations
            if [[ $source_layer == "domain" && $target_layer == "infrastructure" ]]; then
                echo -e "${RED}❌ VIOLATION: $source -> $target (domain -> infrastructure)${NC}"
                violations=$((violations + 1))
            elif [[ $source_layer == "domain" && $target_layer == "presentation" ]]; then
                echo -e "${RED}❌ VIOLATION: $source -> $target (domain -> presentation)${NC}"
                violations=$((violations + 1))
            elif [[ $source_layer == "application" && $target_layer == "presentation" ]]; then
                echo -e "${RED}❌ VIOLATION: $source -> $target (application -> presentation)${NC}"
                violations=$((violations + 1))
            fi
        fi
    done < /tmp/goforms_deps.txt

    if [ $violations -eq 0 ]; then
        echo -e "${GREEN}✅ No dependency violations found${NC}"
    fi

    return $violations
}

# Function to check for circular dependencies
check_circular_deps() {
    echo -e "\n${YELLOW}🔄 Checking for circular dependencies...${NC}"

    local violations=0

    # Use go mod graph to detect cycles
    cd "$PROJECT_ROOT"
    if go mod graph | grep -E "github.com/goformx/goforms.*github.com/goformx/goforms" | \
       awk '{print $1, $2}' | sort | uniq -d > /tmp/circular_deps.txt 2>/dev/null; then

        if [ -s /tmp/circular_deps.txt ]; then
            echo -e "${RED}❌ CIRCULAR DEPENDENCIES DETECTED:${NC}"
            cat /tmp/circular_deps.txt
            violations=$((violations + 1))
        fi
    fi

    if [ $violations -eq 0 ]; then
        echo -e "${GREEN}✅ No circular dependencies found${NC}"
    fi

    return $violations
}

# Function to validate layer structure
validate_layer_structure() {
    echo -e "\n${YELLOW}🏗️ Validating layer structure...${NC}"

    local violations=0

    # Check if all required layers exist
    for layer in "${LAYERS[@]}"; do
        if [ ! -d "internal/$layer" ]; then
            echo -e "${RED}❌ Missing layer: internal/$layer${NC}"
            violations=$((violations + 1))
        else
            echo -e "${GREEN}✅ Layer exists: internal/$layer${NC}"
        fi
    done

    # Check for proper layer organization
    if [ -d "internal/domain" ]; then
        if [ ! -d "internal/domain/entities" ] && [ ! -d "internal/domain/services" ]; then
            echo -e "${YELLOW}⚠️ Warning: Domain layer should contain entities/ and services/${NC}"
        fi
    fi

    if [ -d "internal/application" ]; then
        if [ ! -d "internal/application/services" ] && [ ! -d "internal/application/dto" ]; then
            echo -e "${YELLOW}⚠️ Warning: Application layer should contain services/ and dto/${NC}"
        fi
    fi

    return $violations
}

# Function to generate dependency report
generate_report() {
    echo -e "\n${YELLOW}📋 Generating architecture report...${NC}"

    local report_file="$PROJECT_ROOT/arch-report.txt"

    {
        echo "GoForms Architecture Report"
        echo "Generated: $(date)"
        echo "=================================="
        echo ""
        echo "Layer Dependencies:"
        echo "-------------------"
        cd "$PROJECT_ROOT"
        go mod graph | grep "github.com/goformx/goforms" | while read -r line; do
            source=$(echo "$line" | cut -d' ' -f1)
            target=$(echo "$line" | cut -d' ' -f2)
            source_layer=$(get_layer_name "$source")
            target_layer=$(get_layer_name "$target")
            echo "$source_layer -> $target_layer ($source -> $target)"
        done
        echo ""
        echo "Layer Statistics:"
        echo "----------------"
        for layer in "${LAYERS[@]}"; do
            if [ -d "internal/$layer" ]; then
                file_count=$(find "internal/$layer" -name "*.go" | wc -l)
                echo "$layer: $file_count Go files"
            fi
        done
    } > "$report_file"

    echo -e "${GREEN}✅ Architecture report generated: $report_file${NC}"
}

# Main execution
main() {
    local total_violations=0

    # Change to project root
    cd "$PROJECT_ROOT"

    # Run all checks
    check_forbidden_imports
    total_violations=$((total_violations + $?))

    analyze_dependencies
    total_violations=$((total_violations + $?))

    check_circular_deps
    total_violations=$((total_violations + $?))

    validate_layer_structure
    total_violations=$((total_violations + $?))

    generate_report

    echo -e "\n${BLUE}📊 ARCHITECTURE LINTING SUMMARY${NC}"
    echo "=================================="

    if [ $total_violations -eq 0 ]; then
        echo -e "${GREEN}🎉 All architecture checks passed!${NC}"
        echo -e "${GREEN}✅ Clean Architecture compliance: 100%${NC}"
        exit 0
    else
        echo -e "${RED}❌ Found $total_violations architecture violations${NC}"
        echo -e "${YELLOW}⚠️ Please fix violations before proceeding${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
