#!/usr/bin/env python3
"""One-shot editorial token migration for remaining surfaces.
Mechanical pass: strip dark: color utils, shadows, backdrop-blur; map
slate/brand/rose/amber/emerald/blue utilities to editorial tokens; map
arbitrary radii to the 4/8/12 scale. Structural anti-patterns (icon-squares,
gradients) are fixed manually afterward.
"""
import re
import sys

FILES = [
    "src/routes/issues/[id]/+page.svelte",
    "src/routes/lapor/+page.svelte",
    "src/routes/stats/+page.svelte",
    "src/routes/admin/+layout.svelte",
    "src/routes/admin/login/+page.svelte",
    "src/routes/admin/issues/+page.svelte",
    "src/routes/admin/issues/[id]/+page.svelte",
    "src/lib/components/IssueHeader.svelte",
    "src/lib/components/IssueGallery.svelte",
    "src/lib/components/IssueStats.svelte",
    "src/lib/components/ShareActions.svelte",
    "src/lib/components/ConsentSheet.svelte",
    "src/lib/components/ImagePicker.svelte",
    "src/lib/components/BrowserPushCard.svelte",
    "src/lib/components/NearbyAlertsPanel.svelte",
    "src/lib/components/NotificationPreferencesPanel.svelte",
    "src/lib/components/PrimaryButton.svelte",
]

VARIANT = r"(?:dark:|hover:|focus:|focus-visible:|group-hover:|active:|disabled:)*"

# (regex, replacement). Order matters; opacity suffix /NN is consumed/dropped.
RAW = [
    # text colors
    (r"text-slate-(?:950|900|800|700)(?:/\d+)?", "text-ink"),
    (r"text-slate-(?:600|500)(?:/\d+)?", "text-muted"),
    (r"text-slate-(?:400|300|200)(?:/\d+)?", "text-subtle"),
    (r"text-brand-(?:400|500|600|700)(?:/\d+)?", "text-brand"),
    (r"text-rose-(?:600|700|800)(?:/\d+)?", "text-brand"),
    (r"text-amber-(?:700|800|900)(?:/\d+)?", "text-muted"),
    (r"text-emerald-(?:600|700|800)(?:/\d+)?", "text-verify-community"),
    (r"text-blue-(?:600|700)(?:/\d+)?", "text-ink"),
    # backgrounds
    (r"bg-slate-(?:950|900)(?:/\d+)?", "bg-ink"),
    (r"bg-slate-(?:100|200)(?:/\d+)?", "bg-sunken"),
    (r"bg-slate-(?:50|500)(?:/\d+)?", "bg-sunken"),
    (r"bg-white(?:/\d+)?", "bg-surface"),
    (r"bg-brand-500(?:/\d+)?", "bg-brand"),
    (r"bg-brand-50(?:/\d+)?", "bg-brand-tint"),
    (r"bg-rose-50(?:/\d+)?", "bg-brand-tint"),
    (r"bg-amber-50(?:/\d+)?", "bg-sunken"),
    (r"bg-emerald-50(?:/\d+)?", "bg-sunken"),
    (r"bg-emerald-500(?:/\d+)?", "bg-verify-community"),
    (r"bg-blue-50(?:/\d+)?", "bg-sunken"),
    # borders
    (r"border-slate-300(?:/\d+)?", "border-hairline-strong"),
    (r"border-slate-(?:100|200)(?:/\d+)?", "border-hairline"),
    (r"border-white(?:/\d+)?", "border-hairline"),
    (r"border-brand-(?:100|200)(?:/\d+)?", "border-brand/30"),
    (r"border-rose-(?:100|200)(?:/\d+)?", "border-brand/30"),
    (r"border-amber-(?:100|200)(?:/\d+)?", "border-hairline"),
    (r"border-emerald-(?:100|200)(?:/\d+)?", "border-hairline"),
    (r"border-blue-(?:100|200)(?:/\d+)?", "border-hairline"),
    # ring
    (r"ring-brand-(?:300|400)(?:/\d+)?", "ring-ink"),
    (r"ring-offset-white", "ring-offset-surface"),
    # gradient stops -> drop to neutral ink/surface
    (r"from-slate-950(?:/\d+)?", "from-ink"),
    (r"via-slate-950(?:/\d+)?", "via-ink"),
    (r"to-slate-950(?:/\d+)?", "to-ink"),
]

# arbitrary radius -> scale
RADII = [
    (r"rounded-(?:t-|b-|tl-|tr-|bl-|br-)?\[(?:2[0-9]|3[0-9]|[4-9][0-9])px\]", "rounded-[12px]"),
    (r"rounded-(?:t-|b-|tl-|tr-|bl-|br-)?\[1[6-9]px\]", "rounded-[8px]"),
    (r"rounded-(?:t-|b-|tl-|tr-|bl-|br-)?\[(?:[4-9]|1[0-5])px\]", "rounded-[4px]"),
    (r"rounded-2xl", "rounded-[8px]"),
    (r"rounded-3xl", "rounded-[12px]"),
]

# kill these utilities entirely (flat editorial)
KILL = [
    r"shadow-\[[^\]]*\]",
    r"shadow-(?:sm|md|lg|xl|2xl|inner|card|soft|brand)\b",
    r"backdrop-blur(?:-[a-z0-9]+)?",
    r"hover:-translate-y-[0-9.]+",
    r"active:translate-y-0",
    r"active:scale-\[[0-9.]+\]",
]


def strip_dark(s: str) -> str:
    # remove dark: color/visual utilities (tokens auto-swap)
    s = re.sub(
        r"dark:(?:hover:|focus:|focus-visible:|group-hover:|active:)*"
        r"(?:text|bg|border|ring|ring-offset|placeholder|from|via|to|divide|decoration|fill|stroke)"
        r"-[a-z]+(?:-\d+)?(?:/\d+)?",
        "",
        s,
    )
    # remove dark: arbitrary values e.g. dark:shadow-[...], dark:bg-[...]
    s = re.sub(r"dark:[a-z-]+-\[[^\]]*\]", "", s)
    return s


def process(text: str):
    n = 0
    text = strip_dark(text)
    for pat in KILL:
        text, c = re.subn(pat, "", text)
        n += c
    for pat, rep in RAW + RADII:
        text, c = re.subn(pat, rep, text)
        n += c
    # collapse class double-spaces inside attributes (cosmetic)
    text = re.sub(r'class="([^"]*)"', lambda m: 'class="' + re.sub(r"\s{2,}", " ", m.group(1)).strip() + '"', text)
    return text, n


def main():
    total = 0
    for f in FILES:
        try:
            with open(f, "r") as fh:
                src = fh.read()
        except FileNotFoundError:
            print(f"SKIP (missing): {f}")
            continue
        out, n = process(src)
        if out != src:
            with open(f, "w") as fh:
                fh.write(out)
        total += n
        print(f"{n:4d}  {f}")
    print(f"TOTAL replacements: {total}")


if __name__ == "__main__":
    main()
