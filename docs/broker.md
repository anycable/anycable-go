publisher (node X) -> (stream message) -> broker -> pubsub -> (broadcast) -> node A
      \                                                         /         | \
       \-----------> (command message)-------> pubsub ---------/          |  \-----> node B
                                                                          \
                                                                           \----> node C
