#!/bin/bash

# Test runner for automation and AI agent system

echo "🧪 Running Automation & AI Agent Test Suite"
echo "=============================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL=0
PASSED=0
FAILED=0

# Function to run tests and track results
run_test() {
    local name=$1
    local package=$2
    local pattern=$3

    echo "📦 Testing: $name"

    if go test "$package" -v -run "$pattern" > /tmp/test_output.txt 2>&1; then
        count=$(grep -c "PASS:" /tmp/test_output.txt || echo "0")
        echo -e "${GREEN}✅ PASSED${NC} ($count tests)"
        PASSED=$((PASSED + count))
        TOTAL=$((TOTAL + count))
    else
        count=$(grep -c "FAIL:" /tmp/test_output.txt || echo "0")
        echo -e "${RED}❌ FAILED${NC} ($count tests)"
        FAILED=$((FAILED + count))
        TOTAL=$((TOTAL + count))

        # Show first error
        echo "   Error details:"
        grep "Error:" /tmp/test_output.txt | head -3 | sed 's/^/   /'
    fi
    echo ""
}

# Run test suites
run_test "Automation Service" "./internal/services" "TestAutomationService"
run_test "Node Processors" "./internal/automation/processors" "Test"
run_test "AI Knowledge Base" "./internal/ai/agent" "Test"

# Summary
echo "=============================================="
echo "📊 Test Summary"
echo "=============================================="
echo -e "Total Tests:  $TOTAL"
echo -e "${GREEN}Passed:       $PASSED${NC}"
echo -e "${RED}Failed:       $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠️  Some tests failed. See details above.${NC}"
    exit 1
fi
