package main

import (
	"PlayerGO/player"
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type grpcServer struct {
	player.UnimplementedPlayerServer
	playlist   *SongsForPlaylist
	cmdChannel *chan int
}

func (s *grpcServer) run() error {
	lis, err := net.Listen("tcp", ":8080")

	if err != nil {
		return err
	}

	srv := grpc.NewServer()

	player.RegisterPlayerServer(srv, s)

	return srv.Serve(lis)
}

func (s *grpcServer) PlaySong(ctx context.Context, req *player.PlaySongRequest,
) (*player.PlaySongResponse, error) {
	resp, err := s.playSong(ctx, req)

	if err != nil {
		log.Println("Cannot handle PlaySong request")
		return nil, err
	}

	return resp, nil
}

func (s *grpcServer) playSong(ctx context.Context, req *player.PlaySongRequest,
) (*player.PlaySongResponse, error) {
	info := s.playlist.playSong(s.cmdChannel)

	return &player.PlaySongResponse{
		Info: info,
	}, nil
}

func (s *grpcServer) PauseSong(ctx context.Context, req *player.PauseSongRequest,
) (*player.PauseSongResponse, error) {
	resp, err := s.pauseSong(ctx, req)

	if err != nil {
		log.Println("Cannot handle PauseSong request")
		return nil, err
	}

	return resp, nil
}

func (s *grpcServer) pauseSong(ctx context.Context, req *player.PauseSongRequest,
) (*player.PauseSongResponse, error) {
	info := s.playlist.pauseSong(s.cmdChannel)

	return &player.PauseSongResponse{
		Info: info,
	}, nil
}

func (s *grpcServer) StopPlay(ctx context.Context, req *player.StopRequest,
) (*player.StopResponse, error) {
	resp, err := s.stopPlay(ctx, req)

	if err != nil {
		log.Println("Cannot handle StopPlay request")
		return nil, err
	}

	return resp, nil
}

func (s *grpcServer) stopPlay(ctx context.Context, req *player.StopRequest,
) (*player.StopResponse, error) {
	info := s.playlist.stopPlay(s.cmdChannel)

	return &player.StopResponse{
		Info: info,
	}, nil
}

func (s *grpcServer) AddSong(ctx context.Context, req *player.AddSongRequest,
) (*player.AddSongResponse, error) {
	resp, err := s.addSong(ctx, req)

	if err != nil {
		log.Println("Cannot handle AddSong request")
		return nil, err
	}

	return resp, nil
}

func (s *grpcServer) addSong(ctx context.Context, req *player.AddSongRequest,
) (*player.AddSongResponse, error) {
	songName := req.GetSongName()
	songDuration := req.GetSongDuration()

	songId, err := s.playlist.addSong(songName, songDuration)

	if err != nil {
		return nil, err
	}

	return &player.AddSongResponse{
		SongId:       int32(songId),
		SongName:     songName,
		SongDuration: songDuration,
	}, nil
}

func (s *grpcServer) DeleteSong(ctx context.Context, req *player.DeleteSongRequest,
) (*player.DeleteSongResponse, error) {
	resp, err := s.deleteSong(ctx, req)

	if err != nil {
		log.Println("Cannot handle DeleteSong request")
		return nil, err
	}

	return resp, nil
}

func (s *grpcServer) deleteSong(ctx context.Context, req *player.DeleteSongRequest,
) (*player.DeleteSongResponse, error) {
	songId := int(req.GetSongId())

	songName, err := s.playlist.deleteSong(songId)

	if err != nil {
		return nil, err
	}

	return &player.DeleteSongResponse{
		SongId:   int32(songId),
		SongName: songName,
	}, nil
}

func (s *grpcServer) ShowSongs(ctx context.Context, req *player.ShowSongsRequest,
) (*player.ShowSongsResponse, error) {
	resp, err := s.showSongs(ctx, req)

	if err != nil {
		log.Println("Cannot handle ShowSongs request")
		return nil, err
	}

	return resp, nil
}

func (s *grpcServer) showSongs(ctx context.Context, req *player.ShowSongsRequest,
) (*player.ShowSongsResponse, error) {
	var songs []*player.Song

	songsFromMain := s.playlist.showSongs()

	for songId, songFromMain := range songsFromMain {
		var song player.Song

		song.SongId = int32(songId) + 1
		song.SongName = songFromMain.SongName
		song.SongDuration = songFromMain.SongDuration.String()

		songs = append(songs, &song)
	}

	return &player.ShowSongsResponse{
		Songs: songs,
	}, nil
}
