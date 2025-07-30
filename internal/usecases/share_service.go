package usecases

import (
	"errors"
	"team-service/internal/entities"
	"team-service/internal/repository"

	"gorm.io/gorm"
)

type ShareService interface {
	ShareFolder(folderID uint, targetUserID, access, ownerID string) error
	RevokeFolderShare(folderID uint, targetUserID, ownerID string) error
	ShareNote(noteID uint, targetUserID, access, ownerID string) error
	RevokeNoteShare(noteID uint, targetUserID, ownerID string) error
	GetTeamAssets(teamID uint) (map[string]interface{}, error)
	GetUserAssets(userID string) (map[string]interface{}, error)
}

type shareService struct {
	shareRepo  repository.ShareRepository
	folderRepo repository.FolderRepository
	noteRepo   repository.NoteRepository
	teamRepo   repository.TeamRepository
	db         *gorm.DB
}

func NewShareService(shareRepo repository.ShareRepository, folderRepo repository.FolderRepository, noteRepo repository.NoteRepository, teamRepo repository.TeamRepository, db *gorm.DB) ShareService {
	return &shareService{
		shareRepo:  shareRepo,
		folderRepo: folderRepo,
		noteRepo:   noteRepo,
		teamRepo:   teamRepo,
		db:         db,
	}
}

func (s *shareService) ShareFolder(folderID uint, targetUserID, access, ownerID string) error {
	if targetUserID == ownerID {
		return errors.New("cannot share folder with yourself")
	}

	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		return errors.New("folder not found")
	}

	if folder.OwnerID != ownerID {
		return errors.New("only the owner can share this folder")
	}

	// Transaction: upsert folder share + upsert note shares
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Handle folder share
		existingShare, err := s.shareRepo.GetFolderShare(folderID, targetUserID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if existingShare != nil {
			existingShare.Access = access
			if err := s.shareRepo.UpdateFolderShare(existingShare); err != nil {
				return err
			}
		} else {
			newShare := &entities.FolderShare{
				FolderID: folderID,
				UserID:   targetUserID,
				Access:   access,
			}
			if err := s.shareRepo.CreateFolderShare(newShare); err != nil {
				return err
			}
		}

		// Share all notes in the folder
		notes, err := s.noteRepo.GetByFolderID(folderID)
		if err != nil {
			return err
		}

		for _, note := range notes {
			existingNoteShare, err := s.shareRepo.GetNoteShare(note.ID, targetUserID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			if existingNoteShare != nil {
				existingNoteShare.Access = access
				if err := s.shareRepo.UpdateNoteShare(existingNoteShare); err != nil {
					return err
				}
			} else {
				newNoteShare := &entities.NoteShare{
					NoteID: note.ID,
					UserID: targetUserID,
					Access: access,
				}
				if err := s.shareRepo.CreateNoteShare(newNoteShare); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (s *shareService) RevokeFolderShare(folderID uint, targetUserID, ownerID string) error {
	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		return errors.New("folder not found")
	}

	if folder.OwnerID != ownerID {
		return errors.New("only the folder owner can revoke access")
	}

	return s.shareRepo.DeleteFolderShare(folderID, targetUserID)
}

func (s *shareService) ShareNote(noteID uint, targetUserID, access, ownerID string) error {
	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		return errors.New("note not found")
	}

	if note.OwnerID != ownerID {
		return errors.New("only owner can share the note")
	}

	existingShare, err := s.shareRepo.GetNoteShare(noteID, targetUserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existingShare != nil {
		existingShare.Access = access
		return s.shareRepo.UpdateNoteShare(existingShare)
	} else {
		newShare := &entities.NoteShare{
			NoteID: noteID,
			UserID: targetUserID,
			Access: access,
		}
		return s.shareRepo.CreateNoteShare(newShare)
	}
}

func (s *shareService) RevokeNoteShare(noteID uint, targetUserID, ownerID string) error {
	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		return errors.New("note not found")
	}

	if note.OwnerID != ownerID {
		return errors.New("only owner can revoke access")
	}

	return s.shareRepo.DeleteNoteShare(noteID, targetUserID)
}

func (s *shareService) GetTeamAssets(teamID uint) (map[string]interface{}, error) {
	userIds, err := s.teamRepo.GetUsersByTeamID(teamID)
	if err != nil {
		return nil, errors.New("failed to fetch team members")
	}

	if len(userIds) == 0 {
		return map[string]interface{}{
			"ownedFolders":  []entities.Folder{},
			"sharedFolders": []entities.Folder{},
			"ownedNotes":    []entities.Note{},
			"sharedNotes":   []entities.Note{},
		}, nil
	}

	var ownedFolders []entities.Folder
	var ownedNotes []entities.Note
	var sharedFolders []entities.Folder
	var sharedNotes []entities.Note

	// Get owned assets
	s.db.Where("owner_id IN ?", userIds).Find(&ownedFolders)
	s.db.Where("owner_id IN ?", userIds).Find(&ownedNotes)

	// Get shared assets
	s.db.Model(&entities.Folder{}).
		Joins("JOIN folder_shares fs ON fs.folder_id = folders.id").
		Where("fs.user_id IN ?", userIds).
		Select("folders.*, fs.access").
		Find(&sharedFolders)

	s.db.Model(&entities.Note{}).
		Joins("JOIN note_shares ns ON ns.note_id = notes.id").
		Where("ns.user_id IN ?", userIds).
		Select("notes.*, ns.access").
		Find(&sharedNotes)

	return map[string]interface{}{
		"ownedFolders":  ownedFolders,
		"sharedFolders": sharedFolders,
		"ownedNotes":    ownedNotes,
		"sharedNotes":   sharedNotes,
	}, nil
}

func (s *shareService) GetUserAssets(userID string) (map[string]interface{}, error) {
	var ownedFolders []entities.Folder
	var sharedFolders []entities.Folder
	var ownedNotes []entities.Note
	var sharedNotes []entities.Note

	// Folders owned by user
	s.db.Where("owner_id = ?", userID).Find(&ownedFolders)

	// Notes owned by user
	s.db.Where("owner_id = ?", userID).Find(&ownedNotes)

	// Folders shared to user
	s.db.Model(&entities.Folder{}).
		Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id = ?", userID).
		Select("folders.*, folder_shares.access").
		Find(&sharedFolders)

	// Notes shared to user
	s.db.Model(&entities.Note{}).
		Joins("JOIN note_shares ON notes.id = note_shares.note_id").
		Where("note_shares.user_id = ?", userID).
		Select("notes.*, note_shares.access").
		Find(&sharedNotes)

	return map[string]interface{}{
		"ownedFolders":  ownedFolders,
		"sharedFolders": sharedFolders,
		"ownedNotes":    ownedNotes,
		"sharedNotes":   sharedNotes,
	}, nil
}
