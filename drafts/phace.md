# [wip] Phace - Drawing around Faces

In an earlier post I detailed some of the Mac photos library format via reverse engineering. The focus was on extracting the facial recognition data. Based on those findings, we'll work through proving that it works in this post. The goal will be to produce a new set of images with any recognized faces highlighted. 

Armed with our current understanding of the data model, the steps are as follows:

 1. Get a path to a `.photoslibrary` directory
 1. Open connections to the internal Person.db and Library.apdb
 1. Select all the photo records from the `RKMaster` table
 1. For each photo select all the faces from the `RKFace` table
 1. Read and decode the associated file into an `image.Image`
 1. Map the face coordinates into pixel space
 1. Draw rectangles (since they're easier than circles) around each face
 1. Save a new image file 
 1. Inspect said file for correctness

* TODO Shoudl 2 and 3 be reversed

Things are working ok, but clearly there are some important
attributes that we skipped around image orientation and adjustments.
I'll tackle those in a follow-up (once I actualy figure them out).
And hopefully also the use of embedding this data in the EXIF tags
which was my real goal. Having the image files themselves truly
be self-contained.

