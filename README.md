# Brogit
More like "Broke Git", the broken version of git.
The idea behind this is to basically work out a way in situations like hackathons to properly "git" everything, while working together, with everyone getting the credit for their work, and not wasting time on merge conflicts, by using the power of the internet, to work live in sync with each other, on a single server


## Internal Monologue
Now, tell me. How would the functions work?? on a larger scale, like an abreviation of our larger tasks...i want to make a git wrapper, that can also sync with all of its users, makes a proper branch for each user, in the live versions, it will combine the lines the user worked on (if multiple users worked on a line, the user with the most characters, and then, the last user), with the dev version, and keep building it. And then, when required, using a command, they can easily merge everything, to make it all work. Maybe like if 2 people work on the same line, when we run the git stuff using brogit, we'll make those changes by pushing in the changes in 1 branch, (toposorted based on the lines sort of ...u get the point), and then, merge 2 of the branches, and then the 2nd users stuff will get added to it, and so on....before all of it, gets merged into dev, and then, they all pull from dev too. This stuff is too wild...idk what's going on really....ight, time to do `git init`.

### Plan Of Action
My work is going to be a layer above git...i'll use git to perform the final tasks. Every user will host the server basically now...sort of decentralized, but when a git commit is requested, it will basically do it individually for everyone. For every user, 
- It will try to recognise the changes that they brought into the world (got to figure this out)
- It will create a new formatted doc, with just the users changes (will timestamp everything), and then, put all the stuff the user worked on in the users branch. (there will be a series of commits and all)
- If someone else also made a change in the same line, ie, the next timestamped entry, it will merge those 2 branches, and then, add the stuff to that users branch...and so on until everything is done perfectly.
- Then brogit will clear out its cache.
- Everything will be pushed, and merged into the `dev` branch.
- So, everyone will get the credit, and the work will be done live, without any merge conflicts...everyone will be working on their own thing....however, if someone goes offline, then, ig there might be a merge conflict, cuz their thing won't match....will figure this out later. Need to first build the basic prototype with the other points.

Im thinking of building this in c++ or golang...not exactly sure
I wrote golang in my resume, but like, it's been a while since i practiced it, so, I don't really have a lot of confidence in it right now.
Meanwhile, c++ also seems like a good choice...also i'll end up learning oop in c++ cuz of it.


## Approach
IDK what I wrote above, but here's what im going to do
1. A central server, will run brogit. It will recieve all the commands to read/write stuff from the clients.
2. I want to do it in GoLang
3. The brogit daemon will store all the diffs from all the users, till the admin runs the commit command. Then, brogit will retrace all the commands, and will make batches/groups of all the commands, that were written by the same user, in the same file. Then, if there's a change in order, it will do all the pushes and pulls, to/from the other users branch, and then, make the changes there. So, it will literally follow things sequentially. In the end, it will push the stuff to the final main branch. 
4. Another simpler thing im thinking of rn, is to have all the users commits in their branch, and once we need to switch, we can push everything to a separate development branch, and then, from there, all the users can keep pulling. But, this could cause a bunch of merge conflicts...maybe...got to think properly. Or, need to keep syncing all the branches...