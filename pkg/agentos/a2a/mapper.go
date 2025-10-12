package a2a

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RunInput represents Agno's internal run input format
// RunInput 表示 Agno 的内部运行输入格式
type RunInput struct {
	Content   string                 `json:"content"`
	Images    []Image                `json:"images,omitempty"`
	Files     []File                 `json:"files,omitempty"`
	Videos    []Video                `json:"videos,omitempty"`
	Audios    []Audio                `json:"audio,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// RunOutput represents Agno's internal run output format
// RunOutput 表示 Agno 的内部运行输出格式
type RunOutput struct {
	Content   string                 `json:"content"`
	Images    []Image                `json:"images,omitempty"`
	Files     []File                 `json:"files,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// Media types
type Image struct {
	URL     string `json:"url,omitempty"`
	Content []byte `json:"content,omitempty"`
}

type File struct {
	Name    string `json:"name,omitempty"`
	Content []byte `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
}

type Video struct {
	URL string `json:"url,omitempty"`
}

type Audio struct {
	URL     string `json:"url,omitempty"`
	Content []byte `json:"content,omitempty"`
}

// MapA2ARequestToRunInput converts A2A request to Agno RunInput
// MapA2ARequestToRunInput 将 A2A 请求转换为 Agno RunInput
func MapA2ARequestToRunInput(req *JSONRPC2Request) (*RunInput, error) {
	msg := &req.Params.Message
	runInput := &RunInput{
		Metadata:  make(map[string]interface{}),
		SessionID: msg.ContextID,
	}

	// Store message metadata
	// 存储消息元数据
	runInput.Metadata["a2a_message_id"] = msg.MessageID
	runInput.Metadata["a2a_agent_id"] = msg.AgentID
	runInput.Metadata["a2a_context_id"] = msg.ContextID

	var textParts []string

	// Process each part
	// 处理每个部分
	for _, part := range msg.Parts {
		switch part.Kind {
		case "text":
			if part.Text != nil {
				textParts = append(textParts, *part.Text)
			}

		case "file":
			if part.File == nil {
				continue
			}

			// Handle file based on MIME type
			// 根据 MIME 类型处理文件
			mimeType := strings.ToLower(part.File.MimeType)

			// Image files
			// 图片文件
			if strings.HasPrefix(mimeType, "image/") {
				img, err := processImageFile(part.File)
				if err != nil {
					return nil, fmt.Errorf("failed to process image: %w", err)
				}
				runInput.Images = append(runInput.Images, *img)
				continue
			}

			// Audio files
			// 音频文件
			if strings.HasPrefix(mimeType, "audio/") {
				audio, err := processAudioFile(part.File)
				if err != nil {
					return nil, fmt.Errorf("failed to process audio: %w", err)
				}
				runInput.Audios = append(runInput.Audios, *audio)
				continue
			}

			// Video files
			// 视频文件
			if strings.HasPrefix(mimeType, "video/") {
				video, err := processVideoFile(part.File)
				if err != nil {
					return nil, fmt.Errorf("failed to process video: %w", err)
				}
				runInput.Videos = append(runInput.Videos, *video)
				continue
			}

			// Generic files
			// 通用文件
			file, err := processGenericFile(part.File)
			if err != nil {
				return nil, fmt.Errorf("failed to process file: %w", err)
			}
			runInput.Files = append(runInput.Files, *file)

		case "data":
			if part.Data != nil {
				// Treat data parts as text content
				// 将数据部分视为文本内容
				textParts = append(textParts, part.Data.Content)
			}
		}
	}

	// Combine all text parts
	// 合并所有文本部分
	runInput.Content = strings.Join(textParts, "\n")

	return runInput, nil
}

// MapRunOutputToTask converts Agno RunOutput to A2A Task
// MapRunOutputToTask 将 Agno RunOutput 转换为 A2A Task
func MapRunOutputToTask(output *RunOutput, inputMsg *Message) *Task {
	task := &Task{
		ID:        generateTaskID(),
		ContextID: inputMsg.ContextID,
		Status:    TaskStatusCompleted,
		History:   []Message{},
	}

	// Build response message
	// 构建响应消息
	responseMsg := Message{
		MessageID: generateMessageID(),
		Role:      "agent",
		AgentID:   inputMsg.AgentID,
		ContextID: inputMsg.ContextID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Parts:     []Part{},
	}

	// Add text content if present
	// 如果存在则添加文本内容
	if output.Content != "" {
		text := output.Content
		responseMsg.Parts = append(responseMsg.Parts, Part{
			Kind: "text",
			Text: &text,
		})
	}

	// Add images as file parts
	// 将图片作为文件部分添加
	for _, img := range output.Images {
		if img.URL != "" {
			uri := img.URL
			responseMsg.Parts = append(responseMsg.Parts, Part{
				Kind: "file",
				File: &FilePart{
					URI:      &uri,
					MimeType: "image/png",
				},
			})
		} else if len(img.Content) > 0 {
			encoded := base64.StdEncoding.EncodeToString(img.Content)
			responseMsg.Parts = append(responseMsg.Parts, Part{
				Kind: "file",
				File: &FilePart{
					Bytes:    &encoded,
					MimeType: "image/png",
				},
			})
		}
	}

	// Add files as file parts
	// 将文件作为文件部分添加
	for _, file := range output.Files {
		if file.URL != "" {
			uri := file.URL
			name := file.Name
			responseMsg.Parts = append(responseMsg.Parts, Part{
				Kind: "file",
				File: &FilePart{
					URI:      &uri,
					Name:     &name,
					MimeType: "application/octet-stream",
				},
			})
		} else if len(file.Content) > 0 {
			encoded := base64.StdEncoding.EncodeToString(file.Content)
			name := file.Name
			responseMsg.Parts = append(responseMsg.Parts, Part{
				Kind: "file",
				File: &FilePart{
					Bytes:    &encoded,
					Name:     &name,
					MimeType: "application/octet-stream",
				},
			})
		}
	}

	// Add original input message to history
	// 将原始输入消息添加到历史
	task.History = append(task.History, *inputMsg)

	// Add response message to history
	// 将响应消息添加到历史
	task.History = append(task.History, responseMsg)

	return task
}

// Helper functions for processing different file types
// 处理不同文件类型的辅助函数

func processImageFile(file *FilePart) (*Image, error) {
	img := &Image{}

	if file.URI != nil {
		img.URL = *file.URI
	} else if file.Bytes != nil {
		decoded, err := base64.StdEncoding.DecodeString(*file.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 image: %w", err)
		}
		img.Content = decoded
	}

	return img, nil
}

func processAudioFile(file *FilePart) (*Audio, error) {
	audio := &Audio{}

	if file.URI != nil {
		audio.URL = *file.URI
	} else if file.Bytes != nil {
		decoded, err := base64.StdEncoding.DecodeString(*file.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 audio: %w", err)
		}
		audio.Content = decoded
	}

	return audio, nil
}

func processVideoFile(file *FilePart) (*Video, error) {
	video := &Video{}

	if file.URI != nil {
		video.URL = *file.URI
	}
	// Note: Video bytes not supported in current RunInput
	// 注意: 当前 RunInput 不支持视频字节

	return video, nil
}

func processGenericFile(file *FilePart) (*File, error) {
	f := &File{}

	if file.Name != nil {
		f.Name = *file.Name
	}

	if file.URI != nil {
		f.URL = *file.URI
	} else if file.Bytes != nil {
		decoded, err := base64.StdEncoding.DecodeString(*file.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 file: %w", err)
		}
		f.Content = decoded
	}

	return f, nil
}

// Utility functions
// 工具函数

func generateTaskID() string {
	return "task-" + uuid.New().String()
}

func generateMessageID() string {
	return "msg-" + uuid.New().String()
}
