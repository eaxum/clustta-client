# YC Application - Clustta


**What is your company going to make? Please describe your product and what it does or will do.**

Clustta is an open-source (AGPL v3), distributed version control and collaboration platform for creative workflows - like GitHub, but built for large binary files used in animation, VFX, game development, and design. Our desktop client lets artists version-control files (3D scenes, textures, video), create checkpoints without duplicating storage (via content-defined chunking and deduplication), organize assets with dependency tracking, and collaborate through studios with role-based access, task assignment, and conflict resolution. Unlike Git/Git LFS, which bolt binary support onto a text-oriented system, Clustta is designed from the ground up for creative files - with efficient delta storage, thumbnail previews, workflow templates, and a UI built for artists, not engineers. Being open-source lets studios audit the code, self-host with full control over their IP, and lets the community drive adoption across creative tools and pipelines.


**Which category best applies to your company?**

Developer Tools / Open Source. More specifically: open-source version control and collaboration infrastructure for creative industries.


**How do or will you make money? How much could you make?**

Open-core model - the client and core server are open-source (AGPL v3), with revenue from managed hosting and enterprise features:

Clustta Cloud (managed hosting) - free tier for individuals, paid plans for studios wanting hosted sync/collaboration without running their own servers. Per-seat + storage-tier pricing ($10-50/seat/month).

Enterprise licenses - premium features for large studios: SSO/SAML, advanced audit logs, priority support, SLAs, dedicated infrastructure. The AGPL license naturally encourages enterprises to purchase commercial licenses rather than open-source their proprietary pipeline integrations.

Talent marketplace - studios using Clustta naturally build collaboration history (who worked on what, in which role, with what tools). Enterprise subscribers get access to a talent platform where studios can source artists with specific skillsets, and where artists can showcase their experience, skills, and portfolios. This turns Clustta from a tool into a platform with built-in network effects.

Support and services - paid onboarding, migration from Perforce/Shotgrid, and custom integration work.

This is the same playbook as GitLab, Sentry, and Supabase - open-source core drives adoption, cloud hosting captures revenue, and the talent marketplace adds a network-effect layer that none of those comparisons have.

Market size: The global digital content creation market is ~$30B+ and growing. Current solutions (Perforce ~$800/seat/yr, Shotgrid, ftrack) are expensive and closed. Even capturing a small segment of the ~2M+ professional 3D artists and studios globally represents a $200M-$1B+ opportunity - and open-source dramatically lowers the barrier to entry.


**If you are applying with the same idea as a previous batch, did anything change? If you applied with a different idea, why did you pivot and what did you learn from the last idea?**

Same idea, but two significant things changed since our last application:

First, we switched to open source (AGPL v3) after feedback from existing and potential customers. We also reconsidered the nature of the industry we are targeting. Studios working on unreleased games, films, and commercial IP care deeply about privacy, transparency, and control over their data. Open source gives them the ability to audit the code, self-host, and verify that nothing leaves their network without their knowledge. For studios working on titles like Fortnite or unreleased feature films, that level of trust is not optional - it is a requirement. Going open source also dramatically widened our top-of-funnel and aligned us with the creative open-source community (Blender, Godot, etc.) where our early adopters live.

Second, we discovered a new revenue opportunity from our existing users. Several studios we onboarded kept coming back asking if we "knew anyone" with a specific skillset - a rigger, a texture artist, a lighting TD. We realized we were sitting on a natural network effect. So we built a talent platform on the backend where studios can source artists with specific skillsets, and where artists can display their previous experience, skills, and links to their portfolios. Studios with an enterprise license/subscription will have access to this marketplace. It turns Clustta from a tool into a platform - the collaboration data we already have (who worked on what, in which role, with what tools) naturally feeds a talent discovery layer.


**Why did you pick this idea to work on? Do you have domain expertise in this area? How do you know people need what you're making?**

We have run Eaxum, an animation studio, for 8 years. Over that time we had to figure out production infrastructure from first principles - building Blender addons, assembling and reassembling animation pipelines year after year, and maintaining those same pipelines for other studios we worked with. Every year brought new Blender versions, new Python dependency headaches, and the constant pain of getting tools to work across the dozens of Windows and Mac configurations that artists at different studios happened to have. We eventually realized we were spending more time fighting the pipeline than doing creative work, and that every studio we touched had the same problem.

That experience led directly to Clustta. We needed a version control and collaboration tool that did not require artists to install Python, configure CLI tools, or understand Git. It had to be a standalone desktop application that just works - across platforms, across studios, with no dependency management.

We also know this is not just our problem. Across forums, Reddit threads, and artist communities, creatives routinely ask for alternatives to Git, SVN, and Perforce that actually work for large binary files. Blender Studio themselves published a detailed benchmark comparing Git LFS, SVN, and Mercurial for production use (https://studio.blender.org/blog/benchmarking-version-control-git-lfs-svn-mercurial/), concluding that none of the existing options are a great fit for creative workflows. Francesco Siddi, the current Blender CEO, has taken a keen interest in Clustta after seeing what we are building and Blender Studio is considering adopting it - a strong signal that one of the largest open-source creative organizations in the world recognizes the gap we are filling.


**Who are your competitors? What do you understand about your business that they don't?**

Git / Git LFS - Not designed for large binaries. LFS is a bolted-on workaround with poor deduplication, no visual previews, and requires CLI literacy that most artists don't have.

SVN (Subversion) - Centralized-only with no offline workflow. Poor handling of large binary files, no content-defined chunking or deduplication, and a developer-oriented interface that alienates creative professionals.

Perforce (Helix Core) - Expensive ($800+/seat/yr), closed-source, complex to administer, centralized-only, and has a terrible UX for non-engineers.

Shotgrid / ftrack - Production tracking tools, not true version control. No chunked deduplication, no offline-first capability, no file recovery. Closed-source and vendor-locked.

Google Drive / Dropbox - No real versioning, no deduplication, no creative-aware features. Closed ecosystems with no self-hosting option.

What we understand: (1) Creative professionals need version control but won't adopt developer tools - the product must be visual-first with thumbnails and statuses, not commit hashes and CLI commands. (2) Storage efficiency via content-defined chunking is non-negotiable when versioning 10GB+ Blender/Maya files. (3) Studios handling sensitive IP (unreleased films, games) demand self-hosted and auditable infrastructure - open-source is the only credible answer to "where does my data go?" (4) The creative tools ecosystem is fragmented; an open-source platform invites community-built integrations (Blender, Unreal, Maya plugins) that no closed vendor can match in breadth. The AGPL license simultaneously drives adoption and protects against cloud competitors free-riding on the codebase without contributing back.


**What tech stack are you using, or planning to use, to build this product?**

Desktop client - Go 1.25 + Wails v3 (native webview), Vue 3 + Vite + Pinia

Database - SQLite (per-project .clst files - portable, offline-first, no server dependency)

Serialization - Protocol Buffers (efficient structured data transfer)

Chunking and deduplication - FastCDC (content-defined chunking), SHA-256 hashing, Zstandard compression

Server - Go (clustta-server for auth, clustta-studio for team management) - both open-source

File watching - fsnotify (cross-platform filesystem events)

Auth - OS keyring integration (go-keyring), JWT-based auth

Platforms - Windows (MSIX), macOS (App Store), Linux (AppImage/nfpm)

License - AGPL v3 (all components)

AI tools - GitHub Copilot for development acceleration
