#!/usr/bin/env bash
set -euo pipefail

# ─── helpers ────────────────────────────────────────────────────────────────

ask() {
    local prompt="$1"
    local var_name="$2"
    local default="${3:-}"

    if [[ -n "$default" ]]; then
        read -rp "$prompt [$default]: " value
        echo "${value:-$default}"
    else
        local value=""
        while [[ -z "$value" ]]; do
            read -rp "$prompt: " value
            if [[ -z "$value" ]]; then
                echo "  (required — please enter a value)" >&2
            fi
        done
        echo "$value"
    fi
}

confirm_overwrite() {
    local file="$1"
    if [[ -f "$file" ]]; then
        read -rp "$file already exists. Overwrite? [y/N] " answer
        [[ "$answer" =~ ^[Yy]$ ]]
    else
        return 0
    fi
}

print_header() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  $1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# ─── .env ───────────────────────────────────────────────────────────────────

print_header "Step 1 — Splitwise API key (.env)"

echo ""
echo "Get your key at: https://secure.splitwise.com/apps"
echo ""

if confirm_overwrite ".env"; then
    api_key=$(ask "SPLITWISE_API_KEY" api_key)
    printf 'SPLITWISE_API_KEY=%s\n' "$api_key" > .env
    echo "  → .env written."
else
    echo "  → .env unchanged."
fi

# ─── config.yaml ────────────────────────────────────────────────────────────

print_header "Step 2 — Group and user IDs (config.yaml)"

echo ""
echo "Tip: your user ID is printed at startup when you run the tool,"
echo "     or fetch it with:"
echo "       curl -H 'Authorization: Bearer <API_KEY>' \\"
echo "            https://secure.splitwise.com/api/v3.0/get_current_user"
echo ""

write_config=true
if ! confirm_overwrite "config.yaml"; then
    write_config=false
fi

if [[ "$write_config" == true ]]; then
    group_id=$(ask "Splitwise group_id (number in the group URL)" group_id)
    my_user_id=$(ask "Your Splitwise user ID" my_user_id)

    # ── members ──────────────────────────────────────────────────────────────

    print_header "Step 3 — Members"

    echo ""
    echo "Add each household member (yourself first, then your partner/s)."
    echo "default_entity is the bank entity used when no [ENTITY] tag is present"
    echo "in the expense notes (e.g. SCOTIA, ITAU-D, AMEX)."
    echo ""

    members_yaml=""
    member_count=0

    while true; do
        member_count=$((member_count + 1))
        echo "── Member $member_count ──"

        sid=$(ask "  Splitwise user ID" sid)
        entity=$(ask "  Default entity (e.g. SCOTIA)" entity)

        members_yaml+="  - splitwise_id: ${sid}\n    default_entity: ${entity}\n"

        echo ""
        read -rp "Add another member? [y/N] " more
        [[ "$more" =~ ^[Yy]$ ]] || break
        echo ""
    done

    # ── write file ────────────────────────────────────────────────────────────

    printf 'splitwise:\n  group_id: %s\n  my_user_id: %s\n\nmembers:\n%b' \
        "$group_id" "$my_user_id" "$members_yaml" > config.yaml

    echo ""
    echo "  → config.yaml written."
fi

# ─── done ───────────────────────────────────────────────────────────────────

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Setup complete."
echo ""
echo "  Run the reconciler with:"
echo "    make run            # current month"
echo "    make run MONTH=YYYY-MM"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
